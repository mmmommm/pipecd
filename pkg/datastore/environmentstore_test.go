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

func TestAddEnvironment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testcases := []struct {
		name        string
		environment *model.Environment
		dsFactory   func(*model.Environment) DataStore
		wantErr     bool
	}{
		{
			name:        "Invalid environment",
			environment: &model.Environment{},
			dsFactory:   func(d *model.Environment) DataStore { return nil },
			wantErr:     true,
		},
		{
			name: "Valid environment",
			environment: &model.Environment{
				Id:        "id",
				Name:      "name",
				ProjectId: "project-id",

				CreatedAt: 1,
				UpdatedAt: 1,
			},
			dsFactory: func(d *model.Environment) DataStore {
				ds := NewMockDataStore(ctrl)
				ds.EXPECT().Create(gomock.Any(), "Environment", d.Id, d)
				return ds
			},
			wantErr: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewEnvironmentStore(tc.dsFactory(tc.environment))
			err := s.AddEnvironment(context.Background(), tc.environment)
			assert.Equal(t, tc.wantErr, err != nil)
		})
	}
}

func TestGetEnvironment(t *testing.T) {
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
					Get(gomock.Any(), "Environment", "id", &model.Environment{}).
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
					Get(gomock.Any(), "Environment", "id", &model.Environment{}).
					Return(fmt.Errorf("err"))
				return ds
			}(),
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewEnvironmentStore(tc.ds)
			_, err := s.GetEnvironment(context.Background(), tc.id)
			assert.Equal(t, tc.wantErr, err != nil)
		})
	}
}

func TestListEnvironment(t *testing.T) {
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
					Next(&model.Environment{}).
					Return(ErrIteratorDone)

				ds := NewMockDataStore(ctrl)
				ds.EXPECT().
					Find(gomock.Any(), "Environment", ListOptions{}).
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
					Next(&model.Environment{}).
					Return(fmt.Errorf("err"))

				ds := NewMockDataStore(ctrl)
				ds.EXPECT().
					Find(gomock.Any(), "Environment", ListOptions{}).
					Return(it, nil)
				return ds
			}(),
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewEnvironmentStore(tc.ds)
			_, err := s.ListEnvironments(context.Background(), tc.opts)
			assert.Equal(t, tc.wantErr, err != nil)
		})
	}
}
