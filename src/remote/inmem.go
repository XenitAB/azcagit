package remote

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers"
	"github.com/xenitab/azcagit/src/config"
)

type inMemRemoteActions int

const (
	InMemRemoteActionsCreate inMemRemoteActions = iota
	InMemRemoteActionsUpdate
	InMemRemoteActionsDelete
)

type inMemRemoteAction struct {
	Name   string
	Action inMemRemoteActions
}

type InMemRemote struct {
	getResponse struct {
		firstRemoteApps  *RemoteApps
		secondRemoteApps *RemoteApps
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
	actions []inMemRemoteAction
}

var _ Remote = (*InMemRemote)(nil)

func NewInMemRemote(cfg config.Config) (*InMemRemote, error) {
	return &InMemRemote{}, nil
}

func (r *InMemRemote) Get(ctx context.Context) (*RemoteApps, error) {
	if !r.getResponse.second {
		r.getResponse.second = true
		return r.getResponse.firstRemoteApps, r.getResponse.firstErr
	} else {
		r.getResponse.second = false
		return r.getResponse.secondRemoteApps, r.getResponse.secondErr
	}
}

func (r *InMemRemote) GetFirstResponse(remoteApps *RemoteApps, err error) {
	r.getResponse.firstRemoteApps = remoteApps
	r.getResponse.firstErr = err
}

func (r *InMemRemote) GetSecondResponse(remoteApps *RemoteApps, err error) {
	r.getResponse.secondRemoteApps = remoteApps
	r.getResponse.secondErr = err
}

func (r *InMemRemote) ResetGetSecond() {
	r.getResponse.second = false
}

func (r *InMemRemote) Create(ctx context.Context, name string, app armappcontainers.ContainerApp) error {
	r.actions = append(r.actions, inMemRemoteAction{Name: name, Action: InMemRemoteActionsCreate})
	return r.createResponse.err
}

func (r *InMemRemote) CreateResponse(err error) {
	r.createResponse.err = err
}

func (r *InMemRemote) Update(ctx context.Context, name string, app armappcontainers.ContainerApp) error {
	r.actions = append(r.actions, inMemRemoteAction{Name: name, Action: InMemRemoteActionsUpdate})
	return r.updateResponse.err
}

func (r *InMemRemote) UpdateResponse(err error) {
	r.updateResponse.err = err
}

func (r *InMemRemote) Delete(ctx context.Context, name string) error {
	r.actions = append(r.actions, inMemRemoteAction{Name: name, Action: InMemRemoteActionsDelete})
	return r.deleteResponse.err
}

func (r *InMemRemote) DeleteResponse(err error) {
	r.deleteResponse.err = err
}

func (r *InMemRemote) Actions() []inMemRemoteAction {
	return r.actions
}

func (r *InMemRemote) ResetActions() {
	r.actions = []inMemRemoteAction{}
}
