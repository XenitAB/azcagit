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
	azureRegion           string
}

var _ Metrics = (*AzureMetrics)(nil)

func NewAzureMetrics(cfg config.ReconcileConfig, credential azcore.TokenCredential) *AzureMetrics {
	// The `//` in `https://monitoring.azure.com//.default` is intentional and the required audience is `https://monitoring.azure.com/`,
	// right now something happens inside of the `runtime` which makes the audience `https://monitoring.azure.com` when there's a single `/`.
	authPolicy := runtime.NewBearerTokenPolicy(credential, []string{"https://monitoring.azure.com//.default"}, nil)
	pl := runtime.NewPipeline("azcagit", "undefined", runtime.PipelineOptions{PerRetry: []policy.Policy{authPolicy}}, &policy.ClientOptions{})
	return &AzureMetrics{
		pl:                    pl,
		customMetricsEndpoint: generateCustomMetricsEndpoint(cfg),
		azureRegion:           sanitizeAzureLocation(cfg.Location),
	}
}

func generateCustomMetricsEndpoint(cfg config.ReconcileConfig) string {
	azureRegion := sanitizeAzureLocation(cfg.Location)
	resourceId := fmt.Sprintf("subscriptions/%s/resourceGroups/%s/providers/Microsoft.App/jobs/%s", cfg.SubscriptionID, cfg.OwnResourceGroupName, cfg.OwnContainerJobName)
	return fmt.Sprintf("https://%s.monitoring.azure.com/%s/metrics", azureRegion, resourceId)

}

func sanitizeAzureLocation(location string) string {
	locationWithoutSpaces := strings.ReplaceAll(location, " ", "")
	lowercaseLocation := strings.ToLower(locationWithoutSpaces)
	return lowercaseLocation
}

func (m *AzureMetrics) Int(ctx context.Context, metricName string, metric int) error {
	customMetrics := newCustomMetrics(m.azureRegion, metricName, float64(metric))
	return m.create(ctx, customMetrics)
}

func (m *AzureMetrics) Duration(ctx context.Context, metricName string, metric time.Duration) error {
	customMetrics := newCustomMetrics(m.azureRegion, metricName, metric.Seconds())
	return m.create(ctx, customMetrics)
}

func (m *AzureMetrics) Success(ctx context.Context, metricName string, metric bool) error {
	metricVal := float64(0)
	if metric {
		metricVal = 1
	}
	customMetrics := newCustomMetrics(m.azureRegion, metricName, metricVal)
	return m.create(ctx, customMetrics)
}

func newCustomMetrics(region string, metricName string, metric float64) CustomMetrics {
	return CustomMetrics{
		Time: time.Now(),
		Data: CustomMetricsData{
			BaseData: CustomMetricsBaseData{
				Metric:    metricName,
				Namespace: "azcagit",
				DimNames: []string{
					"region",
				},
				Series: []CustomMetricsSeries{
					{
						DimValues: []string{
							region,
						},
						Min:   metric,
						Max:   metric,
						Sum:   metric,
						Count: 1,
					},
				},
			},
		},
	}
}

type CustomMetricsSeries struct {
	DimValues []string `json:"dimValues,omitempty"`
	Min       float64  `json:"min,omitempty"`
	Max       float64  `json:"max,omitempty"`
	Sum       float64  `json:"sum,omitempty"`
	Count     int      `json:"count,omitempty"`
}

type CustomMetricsBaseData struct {
	Metric    string                `json:"metric,omitempty"`
	Namespace string                `json:"namespace,omitempty"`
	DimNames  []string              `json:"dimNames,omitempty"`
	Series    []CustomMetricsSeries `json:"series,omitempty"`
}

type CustomMetricsData struct {
	BaseData CustomMetricsBaseData `json:"baseData,omitempty"`
}

type CustomMetrics struct {
	Time time.Time         `json:"time,omitempty"`
	Data CustomMetricsData `json:"data,omitempty"`
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
