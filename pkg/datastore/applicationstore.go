// Copyright 2020 The PipeCD Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datastore

import (
	"context"
	"fmt"
	"time"

	"github.com/pipe-cd/pipecd/pkg/model"
)

const ApplicationModelKind = "Application"

var applicationFactory = func() interface{} {
	return &model.Application{}
}

type ApplicationStore interface {
	AddApplication(ctx context.Context, app *model.Application) error
	EnableApplication(ctx context.Context, id string) error
	DisableApplication(ctx context.Context, id string) error
	DeleteApplication(ctx context.Context, id string) error
	GetApplication(ctx context.Context, id string) (*model.Application, error)
	ListApplications(ctx context.Context, opts ListOptions) ([]*model.Application, string, error)
	UpdateApplication(ctx context.Context, id string, updater func(*model.Application) error) error
	PutApplicationSyncState(ctx context.Context, id string, syncState *model.ApplicationSyncState) error
	PutApplicationMostRecentDeployment(ctx context.Context, id string, status model.DeploymentStatus, deployment *model.ApplicationDeploymentReference) error
}

type applicationStore struct {
	backend
	nowFunc func() time.Time
}

func NewApplicationStore(ds DataStore) ApplicationStore {
	return &applicationStore{
		backend: backend{
			ds: ds,
		},
		nowFunc: time.Now,
	}
}

func (s *applicationStore) AddApplication(ctx context.Context, app *model.Application) error {
	now := s.nowFunc().Unix()
	if app.CreatedAt == 0 {
		app.CreatedAt = now
	}
	if app.UpdatedAt == 0 {
		app.UpdatedAt = now
	}
	if err := app.Validate(); err != nil {
		return err
	}
	return s.ds.Create(ctx, ApplicationModelKind, app.Id, app)
}

func (s *applicationStore) EnableApplication(ctx context.Context, id string) error {
	return s.ds.Update(ctx, ApplicationModelKind, id, applicationFactory, func(e interface{}) error {
		app := e.(*model.Application)
		if app.Deleted {
			return fmt.Errorf("cannot enable a deleted application: %w", ErrInvalidArgument)
		}
		app.Disabled = false
		app.UpdatedAt = s.nowFunc().Unix()
		return nil
	})
}

func (s *applicationStore) DisableApplication(ctx context.Context, id string) error {
	return s.ds.Update(ctx, ApplicationModelKind, id, applicationFactory, func(e interface{}) error {
		app := e.(*model.Application)
		if app.Deleted {
			return fmt.Errorf("cannot disable a deleted application: %w", ErrInvalidArgument)
		}
		app.Disabled = true
		app.UpdatedAt = s.nowFunc().Unix()
		return nil
	})
}

func (s *applicationStore) DeleteApplication(ctx context.Context, id string) error {
	return s.ds.Update(ctx, ApplicationModelKind, id, applicationFactory, func(e interface{}) error {
		now := s.nowFunc().Unix()
		app := e.(*model.Application)
		app.Deleted = true
		app.Disabled = true
		app.DeletedAt = now
		app.UpdatedAt = now
		return nil
	})
}

func (s *applicationStore) GetApplication(ctx context.Context, id string) (*model.Application, error) {
	var entity model.Application
	if err := s.ds.Get(ctx, ApplicationModelKind, id, &entity); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (s *applicationStore) ListApplications(ctx context.Context, opts ListOptions) ([]*model.Application, string, error) {
	it, err := s.ds.Find(ctx, ApplicationModelKind, opts)
	if err != nil {
		return nil, "", err
	}
	apps := make([]*model.Application, 0)
	for {
		var app model.Application
		err := it.Next(&app)
		if err == ErrIteratorDone {
			break
		}
		if err != nil {
			return nil, "", err
		}
		apps = append(apps, &app)
	}

	// In case there is no more elements found, cursor should be set to empty too.
	if len(apps) == 0 {
		return apps, "", nil
	}
	cursor, err := it.Cursor()
	if err != nil {
		return nil, "", err
	}
	return apps, cursor, nil
}

func (s *applicationStore) UpdateApplication(ctx context.Context, id string, updater func(*model.Application) error) error {
	now := s.nowFunc().Unix()
	return s.ds.Update(ctx, ApplicationModelKind, id, applicationFactory, func(e interface{}) error {
		a := e.(*model.Application)
		if a.Deleted {
			return fmt.Errorf("cannot update a deleted application: %w", ErrInvalidArgument)
		}
		if err := updater(a); err != nil {
			return err
		}
		a.UpdatedAt = now
		return a.Validate()
	})
}

func (s *applicationStore) PutApplicationSyncState(ctx context.Context, id string, syncState *model.ApplicationSyncState) error {
	return s.UpdateApplication(ctx, id, func(a *model.Application) error {
		a.SyncState = syncState
		return nil
	})
}

func (s *applicationStore) PutApplicationMostRecentDeployment(ctx context.Context, id string, status model.DeploymentStatus, deployment *model.ApplicationDeploymentReference) error {
	return s.UpdateApplication(ctx, id, func(a *model.Application) error {
		switch status {
		case model.DeploymentStatus_DEPLOYMENT_SUCCESS:
			a.MostRecentlySuccessfulDeployment = deployment
		case model.DeploymentStatus_DEPLOYMENT_PENDING:
			a.MostRecentlyTriggeredDeployment = deployment
		}
		return nil
	})
}
