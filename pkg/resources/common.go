package resources

import (
	"github.com/go-logr/logr"

	"github.com/rv-0451/ns-controller/pkg/metrics"
)

func processMetricProviderErrors(err error, reslog logr.Logger) (bool, error) {
	// if metric not found, probably because the cluster is empty and there are no metrics yet
	// in this case it is ok to not return error
	mnferr, ok := err.(*metrics.MetricNotFoundError)
	if ok {
		reslog.Info("Got MetricNotFoundError, probably because the cluster is empty and there are no metrics yet.", "error", mnferr.Error())
		return false, nil
	}
	return false, err
}
