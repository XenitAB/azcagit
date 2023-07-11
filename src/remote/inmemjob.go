package remote

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

type InMemJobActions int

const (
	InMemJobActionsCreate InMemJobActions = iota
	InMemJobActionsUpdate
	InMemJobActionsDelete
)

type InMemJobAction struct {
	Name   string
	Action InMemJobActions
	Job    armappcontainers.Job
}

type InMemJob struct {
	getResponse struct {
		firstRemoteJobs  *RemoteJobs
		secondRemoteJobs *RemoteJobs
		firstErr         error
		secondErr        error
		second           bool
	}
	createResponse struct {
		err error
	}
	updateResponse struct {
		err error
	}
	deleteResponse struct {
		err error
	}
	actions []InMemJobAction
}

var _ Job = (*InMemJob)(nil)

func NewInMemJob() *InMemJob {
	return &InMemJob{}
}

func (r *InMemJob) Get(ctx context.Context) (*RemoteJobs, error) {
	if !r.getResponse.second {
		r.getResponse.second = true
		return r.getResponse.firstRemoteJobs, r.getResponse.firstErr
	} else {
		r.getResponse.second = false
		return r.getResponse.secondRemoteJobs, r.getResponse.secondErr
	}
}

func (r *InMemJob) GetFirstResponse(remoteJobs *RemoteJobs, err error) {
	r.getResponse.firstRemoteJobs = remoteJobs
	r.getResponse.firstErr = err
}

func (r *InMemJob) GetSecondResponse(remoteJobs *RemoteJobs, err error) {
	r.getResponse.secondRemoteJobs = remoteJobs
	r.getResponse.secondErr = err
}

func (r *InMemJob) ResetGetSecond() {
	r.getResponse.second = false
}

func (r *InMemJob) Create(ctx context.Context, name string, job armappcontainers.Job) error {
	r.actions = append(r.actions, InMemJobAction{Name: name, Action: InMemJobActionsCreate, Job: job})
	return r.createResponse.err
}

func (r *InMemJob) CreateResponse(err error) {
	r.createResponse.err = err
}

func (r *InMemJob) Update(ctx context.Context, name string, job armappcontainers.Job) error {
	r.actions = append(r.actions, InMemJobAction{Name: name, Action: InMemJobActionsUpdate, Job: job})
	return r.updateResponse.err
}

func (r *InMemJob) UpdateResponse(err error) {
	r.updateResponse.err = err
}

func (r *InMemJob) Delete(ctx context.Context, name string) error {
	r.actions = append(r.actions, InMemJobAction{Name: name, Action: InMemJobActionsDelete, Job: armappcontainers.Job{}})
	return r.deleteResponse.err
}

func (r *InMemJob) DeleteResponse(err error) {
	r.deleteResponse.err = err
}

func (r *InMemJob) Actions() []InMemJobAction {
	return r.actions
}

func (r *InMemJob) ResetActions() {
	r.actions = []InMemJobAction{}
}
