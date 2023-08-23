package metrics

import (
	"bytes"
	"fmt"
	"io"

	"net/http"
	"net/url"
	"regexp"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"

	"github.com/rv-0451/ns-controller/pkg/utils"
)

const ksmErrorValue float64 = 0

type FilterType int

const (
	Equal FilterType = iota
	NotEqual
	Regexp
)

type KSMFilter struct {
	Type   FilterType
	Fields map[string]string
}

type KubeStateMetrics struct {
	mfMap map[string]*dto.MetricFamily
}

func NewKubeStateMetrics() *KubeStateMetrics {
	in, err := readFromService("http://kube-state-metrics.ns-controller.svc:8081")
	if err != nil {
		metricslog.Error(err, "Failed to read metrics from kubernetes ksm service")
	}

	mfMap, err := parseMF(in)
	if err != nil {
		metricslog.Error(err, "Failed to parse metrics into MetricFamilies")
	}

	return &KubeStateMetrics{
		mfMap: mfMap,
	}
}

func readFromService(serviceUrl string) (io.Reader, error) {
	u, _ := url.ParseRequestURI(serviceUrl)
	u.Path = "metrics"
	urlStr := u.String()

	client := utils.NewHttpClient()
	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Close = true

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(bodyBytes), nil
}

func parseMF(in io.Reader) (map[string]*dto.MetricFamily, error) {
	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(in)
	if err != nil {
		return nil, err
	}
	return mf, nil
}

func (ksm *KubeStateMetrics) getMF(key string) (*dto.MetricFamily, error) {
	mf, ok := ksm.mfMap[key]
	if !ok {
		return nil, &MetricNotFoundError{Msg: fmt.Sprintf("Metric was not found: '%s'", key)}
	}
	return mf, nil
}

func (ksm *KubeStateMetrics) SumMetric(mf *dto.MetricFamily, filters []KSMFilter) (float64, error) {
	metricSum := 0.0
metric:
	for _, m := range mf.Metric {
		for _, l := range m.Label {
			for _, filter := range filters {
				switch filter.Type {
				case Equal:
					for k, v := range filter.Fields {
						if *l.Name == k && *l.Value != v {
							continue metric
						}
					}
				case NotEqual:
					for k, v := range filter.Fields {
						if *l.Name == k && *l.Value == v {
							continue metric
						}
					}
				case Regexp:
					for k, v := range filter.Fields {
						match, err := regexp.MatchString(v, *l.Value)
						if err != nil {
							return ksmErrorValue, err
						}
						if *l.Name == k && !match {
							continue metric
						}
					}
				}
			}
		}
		metricSum += *m.Gauge.Value
	}
	return metricSum, nil
}

func (ksm *KubeStateMetrics) GetAllocatableMemorySum() (float64, error) {
	mf, err := ksm.getMF("kube_node_status_allocatable_memory_bytes")
	if err != nil {
		return ksmErrorValue, err
	}
	filters := []KSMFilter{
		{NotEqual, map[string]string{"node": ""}},
	}
	allocatableMemorySum, err := ksm.SumMetric(mf, filters)
	if err != nil {
		return ksmErrorValue, err
	}
	return allocatableMemorySum, nil
}

func (ksm *KubeStateMetrics) GetPodsLimitSum() (float64, error) {
	mf, err := ksm.getMF("kube_pod_container_resource_limits_memory_bytes")
	if err != nil {
		return ksmErrorValue, err
	}
	filters := []KSMFilter{
		{NotEqual, map[string]string{"node": ""}},
	}
	podsLimitSum, err := ksm.SumMetric(mf, filters)
	if err != nil {
		return ksmErrorValue, err
	}
	return podsLimitSum, nil
}

func (ksm *KubeStateMetrics) QueryMemoryOverprovisioning() (float64, error) {
	podsLimitSum, err := ksm.GetPodsLimitSum()
	if err != nil {
		return ksmErrorValue, err
	}
	metricslog.Info("Sum of all pod memory limits with specific taints:", "podsLimitSum:", podsLimitSum)

	allocatableMemorySum, err := ksm.GetAllocatableMemorySum()
	if err != nil {
		return ksmErrorValue, err
	}
	metricslog.Info("Sum of all allocatable memory on all nodes:", "allocatableMemorySum:", allocatableMemorySum)

	return podsLimitSum / allocatableMemorySum, nil
}
