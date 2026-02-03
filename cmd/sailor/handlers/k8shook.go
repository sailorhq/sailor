package handlers

import (
	"encoding/json"
	"errors"
	"fmt"

	plugrpc "github.com/sailorhq/plug/sdk/proto"
	"github.com/sailorhq/sailor/cmd/sailor/sail"
	"github.com/sailorhq/sailor/internal/bige"
	"github.com/sailorhq/sailor/internal/constants"

	"github.com/valyala/fasthttp"
	"github.com/wI2L/jsondiff"
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

	// before doing anything with this request, we need to know where the Sailor is being hosted
	// if the HostURL in sailor setting is not set then the mutation will not happen, because
	// SailorConsumer will fallback to calling GET /resource API if it does not finds the config
	// inside the volume
	if sc.setting.HostURL == "" {
		sc.Log.Error("admission hook cannot proceed without HostURL set in sailor settings")
		admissionResp.Response.Result = &metav1.Status{
			Code:    fasthttp.StatusBadRequest,
			Message: "no HostURL set in sailor settings",
		}
		respBytes, _ := json.Marshal(admissionResp)
		ctx.SetContentType("application/json")
		ctx.SetBody(respBytes)
		return
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
				jsonPatch := admissionv1.PatchTypeJSONPatch
				admissionResp.Response.PatchType = &jsonPatch
				admissionResp.Response.Patch = patch
			}
		}
	}

	respBytes, _ := json.Marshal(admissionResp)
	ctx.SetContentType("application/json")
	ctx.SetBody(respBytes)
}

func (sc *SailorCore) createSailorPatch(pod *corev1.Pod) ([]byte, error) {
	mutatedPod := pod.DeepCopy()

	addConfigVolume := true
	addSecretVolume := true

	// here we check if sailor volumes are already present .. if yes then
	// we don't mount the volumes
	for _, v := range pod.Spec.Volumes {
		// TODO :: get these names from contants
		switch v.Name {
		case constants.CONFIG_VOLUME_NAME:
			addConfigVolume = false
		case constants.SECRET_VOLUME_NAME:
			addSecretVolume = false
		}
	}

	// TODO :: check if any deployments has been done for this project yet..
	// if not we ignore doing any mutation for this pod and its container
	var (
		err                  error
		containerToOperateOn string
		ns                   string
		projectKey           string
		csail                = sc.SailorSail.(*sail.CoreSail)
		resourceKeys         []string
	)
	// --- CORE MUTATION LOGIC ---
	for i := range mutatedPod.Spec.Containers {
		if _, ok := sc.dbconns[projectKey]; !ok {
			sc.Log.Info("sailor project not found for container",
				zap.String("projectKey", projectKey),
				zap.String("containerName", containerToOperateOn))
			continue
		}

		// ns = mutatedPod.Namespace
		containerToOperateOn = mutatedPod.Spec.Containers[i].Name
		ns = mutatedPod.Namespace
		projectKey = csail.Core_CreateProjectKey(mutatedPod.Namespace, containerToOperateOn)

		// if HostURL is set in sailor settings, we do not need to set any other environment variable other than
		// the URI
		if sc.setting.HostURL != "" {
			uriEnv := corev1.EnvVar{
				Name: "SAILOR_URI",
				Value: createSailorURI(mutatedPod.Namespace,
					containerToOperateOn, sc.setting.AccessKey, sc.setting.SecretKey, sc.setting.HostURL),
			}

			mutatedPod.Spec.Containers[i].Env = append(mutatedPod.Spec.Containers[i].Env, uriEnv)
		}

		resourceKeys, err = csail.GetResourceKeys(projectKey)
		if err != nil {
			return nil, err
		}

		// if there are no resources created in this project then we don't mutate anything
		if len(resourceKeys) == 0 {
			return nil, errors.New("no resource created in this project")
		}

		// for each resources present in this project, we create volume mounts based on their kind
		for _, rk := range resourceKeys {
			switch rk {
			case constants.KindConfig:
				// add volume mounts for sailor config
				mutatedPod.Spec.Containers[i].VolumeMounts = append(
					mutatedPod.Spec.Containers[i].VolumeMounts,
					corev1.VolumeMount{
						Name:      constants.CONFIG_VOLUME_NAME,
						MountPath: constants.CONFIG_VOLUME_PATH,
					})
			case constants.KindSecret:
				// add volume mounts for sailor secret
				mutatedPod.Spec.Containers[i].VolumeMounts = append(
					mutatedPod.Spec.Containers[i].VolumeMounts,
					corev1.VolumeMount{
						Name:      constants.SECRET_VOLUME_NAME,
						MountPath: constants.SECRET_VOLUME_PATH,
					})
			default:
				mutatedPod.Spec.Containers[i].VolumeMounts = append(
					mutatedPod.Spec.Containers[i].VolumeMounts,
					corev1.VolumeMount{
						Name:      constants.MISC_VOLUME_NAME,
						MountPath: constants.MISC_VOLUME_PATH,
					})
			}
		}
	}

	// TODO :: we need to expose a feature where in the developer can mention that this volume is not optional
	// and pod should fail if it is not present and mounted -- this should be in the each resource setting
	sailorVolumeIsOptional := true
	if addConfigVolume && containerToOperateOn != "" {
		mutatedPod.Spec.Volumes = append(mutatedPod.Spec.Volumes, corev1.Volume{
			Name: constants.CONFIG_VOLUME_NAME,
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
			Name: constants.SECRET_VOLUME_NAME,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: fmt.Sprintf("%s-secret", containerToOperateOn),
					Optional:   &sailorVolumeIsOptional,
				},
			},
		})
	}

	// in the end we add a label that sailor injection was successful inside the pod
	mutatedPod.ObjectMeta.Labels[constants.LABEL_ADMISSION] = "ok"
	// --- END CORE MUTATION LOGIC ---

	patch, err := jsondiff.Compare(pod, mutatedPod)
	if err != nil {
		return nil, err
	}

	finalPatchBytes, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	// we are here because patch was generated successfully, this means that nothing can go wrong
	// after this step so we are going to check for release tagging now
	if releaseVersion, ok := pod.Labels[constants.LABEL_RELEASE_TAG]; ok {
		for _, rk := range resourceKeys {
			switch rk {
			case constants.KindConfig:
				configVer, err := csail.GetPinnedVersion(projectKey, rk, releaseVersion)
				if err != nil {
					continue
				}

				// check if the current deployed version and ignore deploying if tagged
				// version is the same
				deployedVer := csail.GetCurrentDeployedVersion(projectKey, rk)
				if deployedVer == configVer {
					// no need to update the ConfigMap in k8s
					continue
				}

				content, _ := csail.BuildResource(
					projectKey,
					rk,
					fmt.Sprintf("%s_version", rk),
					false,
					bige.ByteFromUInt32(configVer))

				err = sc.plugman.FireDeploy(&plugrpc.DeployRequest{
					Ns:          ns,
					App:         containerToOperateOn,
					Kind:        rk,
					ResourceKey: rk,
					Version:     configVer,
					Content:     []byte(content),
				})
				if err != nil {
					return nil, err
				}
			}
		}

	}

	return finalPatchBytes, nil
}
