// sailor
// Copyright (C) 2025 SailorHQ and Ashish Shekar (codekidX)

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	v1 "github.com/codekidx/sailor/pkg/core/v1"
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

	params := extractSailorParams(ctx)

	if params.Ns == "" || params.App == "" || params.Kind == "" {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "namespace or app should not be empty"})
		return
	}

	if !params.Kind.IsOneOf(KindConfig, KindMisc, KindSecret) {
		ctx.SetStatusCode(http.StatusBadRequest)
		enc.Encode(ResponseMessage{Message: "unknown resource kind"})
		return
	}

	if params.Kind.IsMisc() && params.ResourceName == "" {
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

	if _, ok := sc.dbconns[params.ProjectKey]; !ok {
		ctx.SetStatusCode(http.StatusInternalServerError)
		enc.Encode(ResponseMessage{Message: "sailor project was not created"})
		return
	}

	err = sc.dbconns[params.ProjectKey].Update(func(tx *bolt.Tx) error {
		resourceKey := params.Kind.ResourceKey(params.ResourceName)
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
		var resource v1.SailorResource
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
		resourceName := fmt.Sprintf("%s-%s", params.App, resourceKey)
		if resource.Setting.Deploy.K8s {
			switch params.Kind {
			case KindConfig, KindMisc:
				contentKey := fmt.Sprintf("_%s", resourceKey)
				cm := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: params.Ns,
					},
					Data: map[string]string{
						contentKey: buildResource(sc.dbconns[params.ProjectKey], resourceKey, versionKey, false),
					},
				}

				if sc.kube != nil {
					kubeConfigMap := sc.kube.CoreV1().ConfigMaps(params.Ns)
					var err error

					_, err = kubeConfigMap.Get(context.TODO(), resourceName, metav1.GetOptions{})
					// this means that kubernetes cannot get any resource
					if err != nil {
						sc.Log.Info("creating a new k8s config resource", zap.String("name", resourceName))
						_, err = kubeConfigMap.Create(context.TODO(), cm, metav1.CreateOptions{})
					} else {
						sc.Log.Info("updating a new k8s config resource", zap.String("name", resourceName))
						_, err = kubeConfigMap.Update(ctx, cm, metav1.UpdateOptions{})
					}

					if err != nil {
						sc.Log.Error("error while deploying configmap",
							zap.String("ns", params.Ns),
							zap.String("app", params.App),
							zap.String("resource", resourceKey),
							zap.Error(err),
						)
						return err
					}
				} else {
					sc.Log.Error("k8s client uninitialized, cannot deploy config to k8s",
						zap.String("ns", params.Ns),
						zap.String("app", params.App),
						zap.String("resource", resourceKey),
					)
				}
			case KindSecret:
				contentKey := fmt.Sprintf("_%s", resourceKey)
				resStr := buildResource(sc.dbconns[params.ProjectKey], resourceKey, versionKey, false)
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: params.Ns,
					},
					Data: map[string][]byte{
						contentKey: []byte(resStr),
					},
				}

				if sc.kube != nil {
					kubeSecrets := sc.kube.CoreV1().Secrets(params.Ns)

					_, err := kubeSecrets.Get(context.TODO(), resourceName, metav1.GetOptions{})
					if err != nil {
						sc.Log.Info("creating a new k8s secret resource", zap.String("name", resourceName))
						_, err = kubeSecrets.Create(context.TODO(), secret, metav1.CreateOptions{})
					} else {
						sc.Log.Info("updating a new k8s secret resource", zap.String("name", resourceName))
						_, err = kubeSecrets.Update(context.TODO(), secret, metav1.UpdateOptions{})
					}
					if err != nil {
						sc.Log.Error("error while deploying configmap",
							zap.String("ns", params.Ns),
							zap.String("app", params.App),
							zap.String("resource", resourceKey),
							zap.Error(err),
						)
						return err
					}
				} else {
					sc.Log.Error("k8s client uninitialized, cannot deploy config to k8s",
						zap.String("ns", params.Ns),
						zap.String("app", params.App),
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
