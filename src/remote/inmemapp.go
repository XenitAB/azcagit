package remote

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
)

type InMemAppActions int

const (
	InMemAppActionsCreate InMemAppActions = iota
	InMemAppActionsUpdate
	InMemAppActionsDelete
)

type InMemAppAction struct {
	Name   string
	Action InMemAppActions
	App    armappcontainers.ContainerApp
}

type InMemApp struct {
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
	actions []InMemAppAction
}

var _ App = (*InMemApp)(nil)

func NewInMemApp() *InMemApp {
	return &InMemApp{}
}

func (r *InMemApp) Get(ctx context.Context) (*RemoteApps, error) {
	if !r.getResponse.second {
		r.getResponse.second = true
		return r.getResponse.firstRemoteApps, r.getResponse.firstErr
	} else {
		r.getResponse.second = false
		return r.getResponse.secondRemoteApps, r.getResponse.secondErr
	}
}

func (r *InMemApp) GetFirstResponse(remoteApps *RemoteApps, err error) {
	r.getResponse.firstRemoteApps = remoteApps
	r.getResponse.firstErr = err
}

func (r *InMemApp) GetSecondResponse(remoteApps *RemoteApps, err error) {
	r.getResponse.secondRemoteApps = remoteApps
	r.getResponse.secondErr = err
}

func (r *InMemApp) ResetGetSecond() {
	r.getResponse.second = false
}

func (r *InMemApp) Create(ctx context.Context, name string, app armappcontainers.ContainerApp) error {
	r.actions = append(r.actions, InMemAppAction{Name: name, Action: InMemAppActionsCreate, App: app})
	return r.createResponse.err
}

func (r *InMemApp) CreateResponse(err error) {
	r.createResponse.err = err
}

func (r *InMemApp) Update(ctx context.Context, name string, app armappcontainers.ContainerApp) error {
	r.actions = append(r.actions, InMemAppAction{Name: name, Action: InMemAppActionsUpdate, App: app})
	return r.updateResponse.err
}

func (r *InMemApp) UpdateResponse(err error) {
	r.updateResponse.err = err
}

func (r *InMemApp) Delete(ctx context.Context, name string) error {
	r.actions = append(r.actions, InMemAppAction{Name: name, Action: InMemAppActionsDelete, App: armappcontainers.ContainerApp{}})
	return r.deleteResponse.err
}

func (r *InMemApp) DeleteResponse(err error) {
	r.deleteResponse.err = err
}

func (r *InMemApp) Actions() []InMemAppAction {
	return r.actions
}

func (r *InMemApp) ResetActions() {
	r.actions = []InMemAppAction{}
}
