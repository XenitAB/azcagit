package remote

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
	"github.com/xenitab/azcagit/src/config"
)

type AzureJob struct {
	resourceGroup string
	client        *armappcontainers.JobsClient
}

var _ Job = (*AzureJob)(nil)

func NewAzureJob(cfg config.Config, cred azcore.TokenCredential) (*AzureJob, error) {
	client, err := armappcontainers.NewJobsClient(cfg.SubscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	return &AzureJob{
		resourceGroup: cfg.ResourceGroupName,
		client:        client,
	}, nil
}

func (r *AzureJob) Get(ctx context.Context) (*RemoteJobs, error) {
	jobs := make(RemoteJobs)
	pager := r.client.NewListByResourceGroupPager(r.resourceGroup, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, job := range nextResult.Value {
			managed := false
			tag, ok := job.Tags["aca.xenit.io"]
			if ok {
				if *tag == "true" {
					managed = true
				}
			}

			jobs[*job.Name] = RemoteJob{
				job,
				managed,
			}
		}
	}

	return &jobs, nil
}

func (r *AzureJob) Create(ctx context.Context, name string, job armappcontainers.Job) error {
	res, err := r.client.BeginCreateOrUpdate(ctx, r.resourceGroup, name, job, &armappcontainers.JobsClientBeginCreateOrUpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create: %w", err)
	}

	_, err = res.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
		Frequency: 5 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to create: %w", err)
	}

	return nil
}

func (r *AzureJob) Update(ctx context.Context, name string, job armappcontainers.Job) error {
	return r.Create(ctx, name, job)
}

func (r *AzureJob) Delete(ctx context.Context, name string) error {
	res, err := r.client.BeginDelete(ctx, r.resourceGroup, name, &armappcontainers.JobsClientBeginDeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	_, err = res.PollUntilDone(ctx, &runtime.PollUntilDoneOptions{
		Frequency: 5 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	return nil
}
