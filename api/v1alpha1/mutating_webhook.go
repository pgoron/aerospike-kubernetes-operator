package v1alpha1

import (
	"context"
	"net/http"

	// "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// // Defaulter defines functions for setting defaults on resources
// type Defaulter interface {
// 	&AerospikeCluster
// 	Default() error
// }

// // DefaultingWebhookFor creates a new Webhook for Defaulting the provided type.
// func DefaultingWebhookFor(aeroCluster *AerospikeCluster) *admission.Webhook {
// 	return &admission.Webhook{
// 		Handler: &mutatingHandler{aeroCluster: aeroCluster},
// 	}
// }

type mutatingHandler struct {
	// aeroCluster *AerospikeCluster
	decoder *admission.Decoder
}

// var _ DecoderInjector = &mutatingHandler{}

// InjectDecoder injects the decoder into a mutatingHandler.
func (h *mutatingHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

// Handle handles admission requests.
func (h *mutatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	// if h.defaulter == nil {
	// 	panic("defaulter should never be nil")
	// }

	// Get the object in the request
	// obj := h.defaulter.DeepCopyObject().(Defaulter)

	obj := &AerospikeCluster{}
	err := h.decoder.Decode(req, obj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Default the object
	return obj.Default()
	// marshalled, err := json.Marshal(obj)
	// if err != nil {
	// 	return admission.Errored(http.StatusInternalServerError, err)
	// }

	// // Create the patch
	// return admission.PatchResponseFromRaw(req.Object.Raw, marshalled)
}
