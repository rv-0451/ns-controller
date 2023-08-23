package resources

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/rv-0451/ns-controller/pkg/metrics"
	"github.com/rv-0451/ns-controller/pkg/utils"
)

var nsreslog = log.Log.WithName("[ns resource]")

type NamespaceValidator struct {
	Namespace *corev1.Namespace
}

func NewNamespaceValidator(ns *corev1.Namespace) *NamespaceValidator {
	log.FromContext(context.Background())
	return &NamespaceValidator{
		Namespace: ns,
	}
}

func (v *NamespaceValidator) MemoryOverprovisioned() (bool, error) {
	var mp metrics.MetricsProvider = metrics.NewMetricsProvider()
	koef, err := mp.QueryMemoryOverprovisioning()
	if err != nil {
		return processMetricProviderErrors(err, nsreslog)
	}
	nsreslog.Info("Current memory overprovisioning.", "overprovisioning", koef)
	if koef < utils.Overprovisioning() {
		return false, nil
	}
	return true, nil
}
