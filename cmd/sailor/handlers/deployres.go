package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/valyala/fasthttp"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeployResourceRequest struct {
	Version string `json:"version"`
}

func (sc *SailorCore) DeployResourceHandler(ctx *fasthttp.RequestCtx) {
	enc := json.NewEncoder(ctx)

	ns := ctx.UserValue("namespace").(string)
	app := ctx.UserValue("app").(string)
	kind := ctx.UserValue("kind").(string)
	var name string
	if n, ok := ctx.UserValue("name").(string); ok {
		name = n
	}

	if ns == "" || app == "" || kind == "" {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "namespace or app should not be empty"})
		return
	}

	if kind != KindConfig && kind != KindSecret && kind != KindMisc {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "unknown resource kind"})
		return
	}

	if kind == KindMisc && name == "" {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "missing resource name"})
		return
	}

	var deployment DeployResourceRequest
	err := json.Unmarshal(ctx.Request.Body(), &deployment)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	if deployment.Version == "" {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "version is required"})
		return
	}

	projectKey := fmt.Sprintf("%s-%s", ns, app)
	if _, ok := sc.dbconns[projectKey]; !ok {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "sailor project was not created"})
		return
	}

	err = sc.dbconns[projectKey].Update(func(tx *bolt.Tx) error {
		var resourceKey = kind
		if kind == KindMisc {
			resourceKey = name
		}
		versionKey := fmt.Sprintf("%s_version", resourceKey)

		deploymentBucket := tx.Bucket([]byte(BUCKET_DEPLOYMENT)).Bucket([]byte(resourceKey))
		depBytes := deploymentBucket.Get([]byte(deployment.Version))
		if depBytes == nil {
			return fmt.Errorf("no deployment with version: %s", deployment.Version)
		}

		metaBucket := tx.Bucket([]byte(BUCKET_META))
		if err := metaBucket.Put([]byte(versionKey), []byte(deployment.Version)); err != nil {
			return err
		}

		resourceBucket := tx.Bucket([]byte(BUCKET_RESOURCE))
		var resource SailorResource
		resBytes := resourceBucket.Get([]byte(resourceKey))
		if err := json.Unmarshal(resBytes, &resource); err != nil {
			return err
		}
		if resBytes, err = json.Marshal(&resource); err != nil {
			return err
		}

		// here we will get the resource setting and do k8s deployment
		// if k8s is marked true .. if k8s deployment fails then we will
		// rollback all the operations under tx and return error to the user
		resourceName := fmt.Sprintf("%s-%s", app, resourceKey)
		if resource.Setting.Deploy.K8s {
			switch kind {
			case KindConfig, KindMisc:
				cm := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: ns,
					},
					Data: map[string]string{
						"_content": buildResource(sc.dbconns[projectKey], resourceKey, versionKey),
					},
				}

				if sc.kube != nil {
					kubeConfigMap := sc.kube.CoreV1().ConfigMaps(ns)
					var err error

					cmRes, _ := kubeConfigMap.Get(context.TODO(), resourceName, metav1.GetOptions{})
					if cmRes == nil {
						_, err = kubeConfigMap.Create(context.TODO(), cm, metav1.CreateOptions{})
					} else {
						_, err = kubeConfigMap.Update(ctx, cm, metav1.UpdateOptions{})
					}

					if err != nil {
						sc.Log.Error("error while deploying configmap",
							zap.String("ns", ns),
							zap.String("app", app),
							zap.String("resource", resourceKey),
							zap.Error(err),
						)
						return tx.Rollback()
					}
				} else {
					sc.Log.Error("k8s client uninitialized, cannot deploy config to k8s",
						zap.String("ns", ns),
						zap.String("app", app),
						zap.String("resource", resourceKey),
					)
				}
			case KindSecret:
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: ns,
					},
					StringData: map[string]string{
						// TODO :: !!!
					},
				}

				if sc.kube != nil {
					kubeSecrets := sc.kube.CoreV1().Secrets(ns)

					secRes, _ := kubeSecrets.Get(context.TODO(), resourceName, metav1.GetOptions{})
					if secRes == nil {
						_, err = kubeSecrets.Create(context.TODO(), secret, metav1.CreateOptions{})
					} else {
						_, err = kubeSecrets.Update(context.TODO(), secret, metav1.UpdateOptions{})
					}
					if err != nil {
						sc.Log.Error("error while deploying configmap",
							zap.String("ns", ns),
							zap.String("app", app),
							zap.String("resource", resourceKey),
							zap.Error(err),
						)
						return tx.Rollback()
					}
				} else {
					sc.Log.Error("k8s client uninitialized, cannot deploy config to k8s",
						zap.String("ns", ns),
						zap.String("app", app),
						zap.String("resource", resourceKey),
					)
				}
			}
		}
		return nil
	})

	if err != nil {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: err.Error()})
		return
	}

	enc.Encode(ResponseMessage{Message: "ok"})
}
