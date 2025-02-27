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

package livestatereporter

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/pipe-cd/pipecd/pkg/app/piped/livestatestore/kubernetes"
	"github.com/pipe-cd/pipecd/pkg/app/server/service/pipedservice"
	"github.com/pipe-cd/pipecd/pkg/config"
	"github.com/pipe-cd/pipecd/pkg/model"
)

const (
	maxNumEventsPerRequest = 1000
)

type kubernetesReporter struct {
	provider              config.PipedCloudProvider
	appLister             applicationLister
	stateGetter           kubernetes.Getter
	eventIterator         kubernetes.EventIterator
	apiClient             apiClient
	flushInterval         time.Duration
	snapshotFlushInterval time.Duration
	logger                *zap.Logger

	snapshotVersions map[string]model.ApplicationLiveStateVersion
}

func newKubernetesReporter(cp config.PipedCloudProvider, appLister applicationLister, stateGetter kubernetes.Getter, apiClient apiClient, logger *zap.Logger) *kubernetesReporter {
	logger = logger.Named("kubernetes-reporter").With(
		zap.String("cloud-provider", cp.Name),
	)
	return &kubernetesReporter{
		provider:              cp,
		appLister:             appLister,
		stateGetter:           stateGetter,
		eventIterator:         stateGetter.NewEventIterator(),
		apiClient:             apiClient,
		flushInterval:         5 * time.Second,
		snapshotFlushInterval: 10 * time.Minute,
		logger:                logger,
		snapshotVersions:      make(map[string]model.ApplicationLiveStateVersion),
	}
}

func (r *kubernetesReporter) Run(ctx context.Context) error {
	r.logger.Info("start running app live state reporter")

	r.logger.Info("waiting for livestatestore to be ready")
	if err := r.stateGetter.WaitForReady(ctx, 10*time.Minute); err != nil {
		r.logger.Error("livestatestore was unable to be ready in time", zap.Error(err))
		return err
	}

	// Do the first snapshot flushing after the statestore becomes ready.
	r.flushSnapshots(ctx)

	snapshotTicker := time.NewTicker(r.snapshotFlushInterval)
	defer snapshotTicker.Stop()

	ticker := time.NewTicker(r.flushInterval)
	defer ticker.Stop()

L:
	for {
		select {
		case <-snapshotTicker.C:
			r.flushSnapshots(ctx)

		case <-ticker.C:
			r.flushEvents(ctx)

		case <-ctx.Done():
			break L
		}
	}

	r.logger.Info("app live state reporter has been stopped")
	return nil
}

func (r *kubernetesReporter) flushSnapshots(ctx context.Context) error {
	// TODO: In the future, maybe we should apply worker model for this or
	// send multiple application states in one request.
	apps := r.appLister.ListByCloudProvider(r.provider.Name)
	for _, app := range apps {
		state, ok := r.stateGetter.GetKubernetesAppLiveState(app.Id)
		if !ok {
			r.logger.Info(fmt.Sprintf("no app state of kubernetes application %s to report", app.Id))
			continue
		}

		snapshot := &model.ApplicationLiveStateSnapshot{
			ApplicationId: app.Id,
			EnvId:         app.EnvId,
			PipedId:       app.PipedId,
			ProjectId:     app.ProjectId,
			Kind:          app.Kind,
			Kubernetes: &model.KubernetesApplicationLiveState{
				Resources: state.Resources,
			},
			Version: &state.Version,
		}
		snapshot.DetermineAppHealthStatus()
		req := &pipedservice.ReportApplicationLiveStateRequest{
			Snapshot: snapshot,
		}

		if _, err := r.apiClient.ReportApplicationLiveState(ctx, req); err != nil {
			r.logger.Error("failed to report application live state",
				zap.String("application-id", app.Id),
				zap.Error(err),
			)
			continue
		}
		r.snapshotVersions[app.Id] = state.Version
		r.logger.Info(fmt.Sprintf("successfully reported application live state for application: %s", app.Id))
	}
	return nil
}

func (r *kubernetesReporter) flushEvents(ctx context.Context) error {
	events := r.eventIterator.Next(maxNumEventsPerRequest)
	if len(events) == 0 {
		return nil
	}

	filteredEvents := make([]*model.KubernetesResourceStateEvent, 0, len(events))
	for i, event := range events {
		snapshotVersion, ok := r.snapshotVersions[event.ApplicationId]
		if ok && event.SnapshotVersion.IsBefore(snapshotVersion) {
			continue
		}
		filteredEvents = append(filteredEvents, &events[i])
	}
	if len(filteredEvents) == 0 {
		return nil
	}

	req := &pipedservice.ReportApplicationLiveStateEventsRequest{
		KubernetesEvents: filteredEvents,
	}
	if _, err := r.apiClient.ReportApplicationLiveStateEvents(ctx, req); err != nil {
		r.logger.Error("failed to report application live state events",
			zap.Error(err),
		)
		return err
	}

	r.logger.Info(fmt.Sprintf("successfully reported %d events about application live state", len(filteredEvents)))
	return nil
}

func (r *kubernetesReporter) ProviderName() string {
	return r.provider.Name
}
