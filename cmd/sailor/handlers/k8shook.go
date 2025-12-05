package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (sc *SailorCore) K8sAdmissionHookHandler(ctx *fasthttp.RequestCtx) {
	var admissionReq admissionv1.AdmissionReview
	if err := json.Unmarshal(ctx.PostBody(), &admissionReq); err != nil {
		sc.Log.Error("error unmarshalling admission review request", zap.Error(err))
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}

	admissionResp := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: admissionv1.SchemeGroupVersion.String(),
			Kind:       "AdmissionReview",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     admissionReq.Request.UID,
			Allowed: true,
		},
	}

	if admissionReq.Request.Kind.Kind == "Pod" && admissionReq.Request.Operation == admissionv1.Create {
		var pod *corev1.Pod
		if err := json.Unmarshal(admissionReq.Request.Object.Raw, &pod); err != nil {
			sc.Log.Error("error unmarshalling pod", zap.Error(err))
			admissionResp.Response.Result = &metav1.Status{
				Code:    fasthttp.StatusBadRequest,
				Message: "error unmarshalling pod",
			}
		} else {
			patch, err := sc.createSailorPatch(pod)
			if err != nil {
				sc.Log.Error("error creating sailor patch", zap.Error(err))
				admissionResp.Response.Result = &metav1.Status{
					Code:    fasthttp.StatusInternalServerError,
					Message: "error creating sailor patch",
				}
			} else {
				admissionResp.Response.PatchType = func() *admissionv1.PatchType {
					pt := admissionv1.PatchTypeJSONPatch
					return &pt
				}()
				admissionResp.Response.Patch = []byte(base64.StdEncoding.EncodeToString(patch))
			}
		}
	}

	respBytes, _ := json.Marshal(admissionResp)
	ctx.SetContentType("application/json")
	ctx.SetBody(respBytes)
}

func (sc *SailorCore) createSailorPatch(pod *corev1.Pod) ([]byte, error) {
	originalJSON, err := json.Marshal(pod)
	if err != nil {
		return nil, err
	}

	mutatedPod := pod.DeepCopy()

	addConfigVolume := true
	addSecretVolume := true
	for _, v := range pod.Spec.Volumes {
		// TODO :: get these names from contants
		switch v.Name {
		case "sailor-config-volume":
			addConfigVolume = false
		case "sailor-secret-volume":
			addSecretVolume = false
		}
	}

	// TODO :: check if any deployments has been done for this project yet..
	// if not we ignore doing any mutation for this pod and its container
	containerToOperateOn := ""
	// --- CORE MUTATION LOGIC ---
	for i := range mutatedPod.Spec.Containers {
		containerName := mutatedPod.Spec.Containers[i].Name
		projectKey := fmt.Sprintf("%s_%s", mutatedPod.Namespace, mutatedPod.Spec.Containers[i].Name)
		if _, ok := sc.dbconns[projectKey]; !ok {
			sc.Log.Info("project not found for container", zap.String("projectKey", projectKey), zap.String("containerName", containerToOperateOn))
			continue
		}

		containerToOperateOn = containerName

		nsEnv := corev1.EnvVar{
			Name:  "SAILOR_NS",
			Value: mutatedPod.Namespace,
		}
		appEnv := corev1.EnvVar{
			Name:  "SAILOR_APP",
			Value: containerName,
		}
		accessKey := corev1.EnvVar{
			Name:  "SAILOR_ACCESS_KEY",
			Value: sc.setting.AccessKey,
		}
		secretKey := corev1.EnvVar{
			Name:  "SAILOR_SECRET_KEY",
			Value: sc.setting.SecretKey,
		}

		mutatedPod.Spec.Containers[i].Env = append(mutatedPod.Spec.Containers[i].Env, nsEnv, appEnv, accessKey, secretKey)

		// TODO :: for now we will add volume mounts for both config and secrets
		// we need to later look into what resource types are present in the project and only
		// mount those here
		// add volume mounts for sailor config
		mutatedPod.Spec.Containers[i].VolumeMounts = append(mutatedPod.Spec.Containers[i].VolumeMounts, corev1.VolumeMount{
			Name:      "sailor-config-volume", // TODO :: get it from constants
			MountPath: "/etc/sailor",
		})
		// add volume mounts for sailor secret
		mutatedPod.Spec.Containers[i].VolumeMounts = append(mutatedPod.Spec.Containers[i].VolumeMounts, corev1.VolumeMount{
			Name:      "sailor-secret-volume", // TODO :: get it from constants
			MountPath: "/etc/sailor/secret",
		})
	}

	// TODO :: we need to expose a feature where in the developer can mention that this volume is not optional
	// and pod should fail if it is not present and mounted -- this should be in the each resource setting
	sailorVolumeIsOptional := true
	if addConfigVolume && containerToOperateOn != "" {
		mutatedPod.Spec.Volumes = append(mutatedPod.Spec.Volumes, corev1.Volume{
			Name: "sailor-config-volume",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-config", containerToOperateOn),
					},
					Optional: &sailorVolumeIsOptional,
				},
			},
		})
	}
	if addSecretVolume && containerToOperateOn != "" {
		mutatedPod.Spec.Volumes = append(mutatedPod.Spec.Volumes, corev1.Volume{
			Name: "sailor-secret-volume",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: fmt.Sprintf("%s-secret", containerToOperateOn),
					Optional:   &sailorVolumeIsOptional,
				},
			},
		})
	}

	// in the end we add a label that sailor injection was successful inside the pod
	mutatedPod.ObjectMeta.Labels["sailor"] = "injected=true"
	// --- END CORE MUTATION LOGIC ---

	mutatedJSON, err := json.Marshal(mutatedPod)
	if err != nil {
		return nil, err
	}

	// Calculate the difference between the two objects
	patch, err := jsonpatch.CreateMergePatch(originalJSON, mutatedJSON)
	if err != nil {
		return nil, err
	}

	return patch, nil
}
