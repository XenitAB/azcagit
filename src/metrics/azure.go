package metrics

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/xenitab/azcagit/src/config"
)

type AzureMetrics struct {
	pl                    runtime.Pipeline
	customMetricsEndpoint string
	region                string
}

var _ Metrics = (*AzureMetrics)(nil)

func NewAzureMetrics(cfg config.Config, credential azcore.TokenCredential) *AzureMetrics {
	authPolicy := runtime.NewBearerTokenPolicy(credential, []string{"https://monitoring.azure.com//.default"}, nil)
	pl := runtime.NewPipeline("azcustommetrics", "v0.0.1", runtime.PipelineOptions{PerRetry: []policy.Policy{authPolicy}}, &policy.ClientOptions{})
	resourceId := "subscriptions/2a6936a5-fc30-492a-ab19-ec59068b5b96/resourceGroups/rg-dev-we-azcagit-platform/providers/Microsoft.App/containerApps/azcagit"
	return &AzureMetrics{
		pl:                    pl,
		customMetricsEndpoint: fmt.Sprintf("https://%s.monitoring.azure.com/%s/metrics", fixedAzureLocation(cfg.Location), resourceId),
		region:                fixedAzureLocation(cfg.Location),
	}
}

func fixedAzureLocation(location string) string {
	locationWithoutSpaces := strings.ReplaceAll(location, " ", "")
	lowercaseLocation := strings.ToLower(locationWithoutSpaces)
	return lowercaseLocation
}

func (m *AzureMetrics) Float64(ctx context.Context, metricName string, metric float64) error {
	customMetrics := newCustomMetrics(m.region, metricName, metric)
	return m.create(ctx, customMetrics)
}

func (m *AzureMetrics) Int(ctx context.Context, metricName string, metric int) error {
	customMetrics := newCustomMetrics(m.region, metricName, float64(metric))
	return m.create(ctx, customMetrics)
}

func newCustomMetrics(region string, metricName string, metric float64) CustomMetrics {
	return CustomMetrics{
		Time: toPtr(time.Now()),
		Data: &CustomMetricsData{
			BaseData: &CustomMetricsBaseData{
				Metric:    &metricName,
				Namespace: toPtr("azcagit"),
				DimNames: []string{
					"region",
				},
				Series: []CustomMetricsSeries{
					{
						DimValues: []string{
							region,
						},
						Min:   &metric,
						Max:   &metric,
						Sum:   &metric,
						Count: toPtr(1),
					},
				},
			},
		},
	}
}

func toPtr[T any](v T) *T {
	return &v
}

type CustomMetricsSeries struct {
	DimValues []string `json:"dimValues,omitempty"`
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
	Sum       *float64 `json:"sum,omitempty"`
	Count     *int     `json:"count,omitempty"`
}

type CustomMetricsBaseData struct {
	Metric    *string               `json:"metric,omitempty"`
	Namespace *string               `json:"namespace,omitempty"`
	DimNames  []string              `json:"dimNames,omitempty"`
	Series    []CustomMetricsSeries `json:"series,omitempty"`
}

type CustomMetricsData struct {
	BaseData *CustomMetricsBaseData `json:"baseData,omitempty"`
}

type CustomMetrics struct {
	Time *time.Time         `json:"time,omitempty"`
	Data *CustomMetricsData `json:"data,omitempty"`
}

func (client *AzureMetrics) create(ctx context.Context, body CustomMetrics) error {
	req, err := client.customCreateRequest(ctx, body)
	if err != nil {
		return err
	}
	resp, err := client.pl.Do(req)
	if err != nil {
		return err
	}
	if !runtime.HasStatusCode(resp, http.StatusOK) {
		return runtime.NewResponseError(resp)
	}
	return nil
}

func (client *AzureMetrics) customCreateRequest(ctx context.Context, body CustomMetrics) (*policy.Request, error) {
	req, err := runtime.NewRequest(ctx, http.MethodPost, client.customMetricsEndpoint)
	if err != nil {
		return nil, err
	}
	req.Raw().Header["Accept"] = []string{"application/json"}
	return req, runtime.MarshalAsJSON(req, body)
}
