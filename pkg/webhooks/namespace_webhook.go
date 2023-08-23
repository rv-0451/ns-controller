package webhooks

import (
	"context"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	res "github.com/rv-0451/ns-controller/pkg/resources"
)

var nswebhooklog = log.Log.WithName("[ns webhook]")

type NamespaceValidatingHandler struct {
	Client  client.Client
	decoder *admission.Decoder
}

func RegisterNamespaceWebhook(mgr manager.Manager) {
	nswebhooklog.Info("Registering NamespaceValidatingHandler.")
	m := NamespaceValidatingHandler{
		Client: mgr.GetClient(),
	}
	webhookServer := mgr.GetWebhookServer()
	webhookServer.CertDir = "/tmp/k8s-webhook-server/serving-certs"
	webhookServer.Register("/validate-v1-namespace", &webhook.Admission{Handler: &m})
}

func (h *NamespaceValidatingHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *NamespaceValidatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	log.FromContext(ctx)

	if req.Operation != admissionv1.Create {
		return admission.Allowed("")
	}

	ns := &corev1.Namespace{}
	if err := h.decoder.DecodeRaw(req.Object, ns); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	nsv := res.NewNamespaceValidator(ns)
	mo, err := nsv.MemoryOverprovisioned()
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if mo {
		return admission.Denied("Request denied due to memory overprovisioning")
	}

	return admission.Allowed("")
}
