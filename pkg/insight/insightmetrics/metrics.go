// Copyright 2021 The PipeCD Authors.
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

package insightmetrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/pipe-cd/pipecd/pkg/datastore"
	"github.com/pipe-cd/pipecd/pkg/insight/insightstore"
	"github.com/pipe-cd/pipecd/pkg/model"
)

const (
	projectKey = "project"
	appKindKey = "app_kind"
)

type insightMetricsCollector struct {
	insightStore insightstore.Store
	projectStore datastore.ProjectStore

	applicationDesc *prometheus.Desc
}

func NewInsightMetricsCollector(is insightstore.Store, ps datastore.ProjectStore) prometheus.Collector {
	return &insightMetricsCollector{
		insightStore: is,
		projectStore: ps,
		applicationDesc: prometheus.NewDesc(
			"insight_application_total",
			"Number of applications currently controlled by control plane",
			[]string{projectKey, appKindKey},
			nil,
		),
	}
}

func (i *insightMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- i.applicationDesc
}

func (i *insightMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	applicationCounts, err := i.collectApplicationCount()
	if err != nil {
		return
	}

	for proj, apps := range applicationCounts {
		for kind, cnt := range apps {
			ch <- prometheus.MustNewConstMetric(
				i.applicationDesc,
				prometheus.GaugeValue,
				float64(cnt),
				proj,
				kind,
			)
		}
	}
}

// collectApplicationCount returns a map like map[projectID]map[kind](number-of-applications).
func (i *insightMetricsCollector) collectApplicationCount() (map[string]map[string]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	projects, err := i.projectStore.ListProjects(ctx, datastore.ListOptions{})
	if err != nil {
		return nil, err
	}
	data := make(map[string]map[string]int, len(projects))
	for idx := range projects {
		counts, err := i.insightStore.LoadApplicationCounts(ctx, projects[idx].Id)
		if err != nil {
			continue
		}
		data[projects[idx].Id] = groupApplicationCounts(counts.Counts)
	}
	return data, nil
}

// groupApplicationCounts groups the number of applications by kind.
func groupApplicationCounts(counts []model.InsightApplicationCount) map[string]int {
	groups := make(map[string]int, len(counts))
	for _, c := range counts {
		kind := c.Labels[model.InsightApplicationCountLabelKey_KIND.String()]
		groups[kind] = groups[kind] + int(c.Count)
	}
	return groups
}
