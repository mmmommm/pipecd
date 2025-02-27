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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/pipe-cd/pipecd/pkg/model"
)

func TestAddApplication(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testcases := []struct {
		name        string
		application *model.Application
		dsFactory   func(*model.Application) DataStore
		wantErr     bool
	}{
		{
			name:        "Invalid application",
			application: &model.Application{},
			dsFactory:   func(d *model.Application) DataStore { return nil },
			wantErr:     true,
		},
		{
			name: "Valid application",
			application: &model.Application{
				Id:        "id",
				Name:      "name",
				EnvId:     "env-id",
				PipedId:   "piped-id",
				ProjectId: "project-id",
				Kind:      model.ApplicationKind_KUBERNETES,
				GitPath: &model.ApplicationGitPath{
					Repo: &model.ApplicationGitRepository{Id: "id"},
					Path: "path",
				},
				CloudProvider: "cloud-provider",

				CreatedAt: 1,
				UpdatedAt: 1,
			},
			dsFactory: func(d *model.Application) DataStore {
				ds := NewMockDataStore(ctrl)
				ds.EXPECT().Create(gomock.Any(), "Application", d.Id, d)
				return ds
			},
			wantErr: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewApplicationStore(tc.dsFactory(tc.application))
			err := s.AddApplication(context.Background(), tc.application)
			assert.Equal(t, tc.wantErr, err != nil)
		})
	}
}

func TestGetApplication(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testcases := []struct {
		name    string
		id      string
		ds      DataStore
		wantErr bool
	}{
		{
			name: "successful fetch from datastore",
			id:   "id",
			ds: func() DataStore {
				ds := NewMockDataStore(ctrl)
				ds.EXPECT().
					Get(gomock.Any(), "Application", "id", &model.Application{}).
					Return(nil)
				return ds
			}(),
			wantErr: false,
		},
		{
			name: "failed fetch from datastore",
			id:   "id",
			ds: func() DataStore {
				ds := NewMockDataStore(ctrl)
				ds.EXPECT().
					Get(gomock.Any(), "Application", "id", &model.Application{}).
					Return(fmt.Errorf("err"))
				return ds
			}(),
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewApplicationStore(tc.ds)
			_, err := s.GetApplication(context.Background(), tc.id)
			assert.Equal(t, tc.wantErr, err != nil)
		})
	}
}

func TestListApplications(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testcases := []struct {
		name    string
		opts    ListOptions
		ds      DataStore
		wantErr bool
	}{
		{
			name: "iterator done",
			opts: ListOptions{},
			ds: func() DataStore {
				it := NewMockIterator(ctrl)
				it.EXPECT().
					Next(&model.Application{}).
					Return(ErrIteratorDone)

				ds := NewMockDataStore(ctrl)
				ds.EXPECT().
					Find(gomock.Any(), "Application", ListOptions{}).
					Return(it, nil)
				return ds
			}(),
			wantErr: false,
		},
		{
			name: "unexpected error occurred",
			opts: ListOptions{},
			ds: func() DataStore {
				it := NewMockIterator(ctrl)
				it.EXPECT().
					Next(&model.Application{}).
					Return(fmt.Errorf("err"))

				ds := NewMockDataStore(ctrl)
				ds.EXPECT().
					Find(gomock.Any(), "Application", ListOptions{}).
					Return(it, nil)
				return ds
			}(),
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewApplicationStore(tc.ds)
			_, _, err := s.ListApplications(context.Background(), tc.opts)
			assert.Equal(t, tc.wantErr, err != nil)
		})
	}
}
