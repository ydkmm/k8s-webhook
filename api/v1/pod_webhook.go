/*
Copyright 2023 ydkmm.
*/

package v1

import (
	"context"
	"encoding/json"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:verbs=create;update,path=/mutate-core-v1-pod,validating=false,failurePolicy=fail,groups=core,resources=pods,versions=v1,name=vpod.kb.io

// PodSideCarMutate mutate Pods
type PodSidecarMutate struct {
	Client  client.Client
	decoder *admission.Decoder
}

func NewPodSideCarMutate(c client.Client) admission.Handler {
	return &PodSidecarMutate{Client: c}
}

// PodSideCarMutate admits a pod if a specific annotation exists.
func (v *PodSidecarMutate) Handle(ctx context.Context, req admission.Request) admission.Response {
	// TODO

	pod := &corev1.Pod{}

	err := v.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	sidecar := corev1.Container{
		Name:            "nginx",
		Image:           "nginx:1.16",
		ImagePullPolicy: corev1.PullIfNotPresent,
		Ports: []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: 80,
			},
		},
	}

	pod.Spec.Containers = append(pod.Spec.Containers, sidecar)

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

// PodSideCarMutate implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (v *PodSidecarMutate) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
