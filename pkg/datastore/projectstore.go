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
	"time"

	"github.com/pipe-cd/pipecd/pkg/model"
)

const ProjectModelKind = "Project"

var projectFactory = func() interface{} {
	return &model.Project{}
}

type ProjectStore interface {
	AddProject(ctx context.Context, proj *model.Project) error
	UpdateProject(ctx context.Context, id string, updater func(project *model.Project) error) error
	UpdateProjectStaticAdmin(ctx context.Context, id, username, password string) error
	EnableStaticAdmin(ctx context.Context, id string) error
	DisableStaticAdmin(ctx context.Context, id string) error
	UpdateProjectSSOConfig(ctx context.Context, id string, sso *model.ProjectSSOConfig) error
	UpdateProjectRBACConfig(ctx context.Context, id string, sso *model.ProjectRBACConfig) error
	GetProject(ctx context.Context, id string) (*model.Project, error)
	ListProjects(ctx context.Context, opts ListOptions) ([]model.Project, error)
}

type projectStore struct {
	backend
	nowFunc func() time.Time
}

func NewProjectStore(ds DataStore) ProjectStore {
	return &projectStore{
		backend: backend{
			ds: ds,
		},
		nowFunc: time.Now,
	}
}

func (s *projectStore) AddProject(ctx context.Context, proj *model.Project) error {
	now := s.nowFunc().Unix()
	if proj.CreatedAt == 0 {
		proj.CreatedAt = now
	}
	if proj.UpdatedAt == 0 {
		proj.UpdatedAt = now
	}
	if err := proj.Validate(); err != nil {
		return err
	}
	return s.ds.Create(ctx, ProjectModelKind, proj.Id, proj)
}

func (s *projectStore) UpdateProject(ctx context.Context, id string, updater func(project *model.Project) error) error {
	now := s.nowFunc().Unix()
	return s.ds.Update(ctx, ProjectModelKind, id, projectFactory, func(e interface{}) error {
		p := e.(*model.Project)
		if err := updater(p); err != nil {
			return err
		}
		p.UpdatedAt = now
		return p.Validate()
	})
}

// UpdateProjectStaticAdmin updates the static admin user settings.
func (s *projectStore) UpdateProjectStaticAdmin(ctx context.Context, id, username, password string) error {
	return s.UpdateProject(ctx, id, func(p *model.Project) error {
		if p.StaticAdmin == nil {
			p.StaticAdmin = &model.ProjectStaticUser{}
		}
		return p.StaticAdmin.Update(username, password)
	})
}

// EnableStaticAdmin enables static admin login.
func (s *projectStore) EnableStaticAdmin(ctx context.Context, id string) error {
	return s.UpdateProject(ctx, id, func(p *model.Project) error {
		p.StaticAdminDisabled = false
		return nil
	})
}

// DisableStaticAdmin disables static admin login.
func (s *projectStore) DisableStaticAdmin(ctx context.Context, id string) error {
	return s.UpdateProject(ctx, id, func(p *model.Project) error {
		p.StaticAdminDisabled = true
		return nil
	})
}

// UpdateProjectSSOConfig updates project single sign on settings.
func (s *projectStore) UpdateProjectSSOConfig(ctx context.Context, id string, sso *model.ProjectSSOConfig) error {
	return s.UpdateProject(ctx, id, func(p *model.Project) error {
		if p.Sso == nil {
			p.Sso = &model.ProjectSSOConfig{}
		}
		p.Sso.Update(sso)
		return nil
	})
}

// UpdateProjectRBACConfig updates project single sign on settings.
func (s *projectStore) UpdateProjectRBACConfig(ctx context.Context, id string, rbac *model.ProjectRBACConfig) error {
	return s.UpdateProject(ctx, id, func(p *model.Project) error {
		p.Rbac = rbac
		return nil
	})
}

func (s *projectStore) GetProject(ctx context.Context, id string) (*model.Project, error) {
	var entity model.Project
	if err := s.ds.Get(ctx, ProjectModelKind, id, &entity); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (s *projectStore) ListProjects(ctx context.Context, opts ListOptions) ([]model.Project, error) {
	it, err := s.ds.Find(ctx, ProjectModelKind, opts)
	if err != nil {
		return nil, err
	}
	ps := make([]model.Project, 0)
	for {
		var p model.Project
		err := it.Next(&p)
		if err == ErrIteratorDone {
			break
		}
		if err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}
	return ps, nil
}
