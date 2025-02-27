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

package cloudrun

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

type ServiceManifest struct {
	Name string
	u    *unstructured.Unstructured
}

func (m ServiceManifest) SetRevision(name string) error {
	return unstructured.SetNestedField(m.u.Object, name, "spec", "template", "metadata", "name")
}

type RevisionTraffic struct {
	RevisionName string `json:"revisionName"`
	Percent      int    `json:"percent"`
}

func (m ServiceManifest) UpdateTraffic(revisions []RevisionTraffic) error {
	items := []interface{}{}
	for i := range revisions {
		out, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&revisions[i])
		if err != nil {
			return fmt.Errorf("unable to set traffic for object: %w", err)
		}
		items = append(items, out)
	}

	return unstructured.SetNestedSlice(m.u.Object, items, "spec", "traffic")
}

func (m ServiceManifest) UpdateAllTraffic(revision string) error {
	return m.UpdateTraffic([]RevisionTraffic{
		{
			RevisionName: revision,
			Percent:      100,
		},
	})
}

func (m ServiceManifest) YamlBytes() ([]byte, error) {
	return yaml.Marshal(m.u)
}

func (m ServiceManifest) AddLabels(labels map[string]string) {
	if len(labels) == 0 {
		return
	}

	lbls := m.u.GetLabels()
	if lbls == nil {
		m.u.SetLabels(labels)
		return
	}
	for k, v := range labels {
		lbls[k] = v
	}
	m.u.SetLabels(lbls)
}

func loadServiceManifest(path string) (ServiceManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ServiceManifest{}, err
	}
	return ParseServiceManifest(data)
}

func ParseServiceManifest(data []byte) (ServiceManifest, error) {
	var obj unstructured.Unstructured
	if err := yaml.Unmarshal(data, &obj); err != nil {
		return ServiceManifest{}, err
	}

	return ServiceManifest{
		Name: obj.GetName(),
		u:    &obj,
	}, nil
}

func DecideRevisionName(sm ServiceManifest, commit string) (string, error) {
	tag, err := FindImageTag(sm)
	if err != nil {
		return "", err
	}
	tag = strings.ReplaceAll(tag, ".", "")

	if len(commit) > 7 {
		commit = commit[:7]
	}
	return fmt.Sprintf("%s-%s-%s", sm.Name, tag, commit), nil
}

func FindImageTag(sm ServiceManifest) (string, error) {
	containers, ok, err := unstructured.NestedSlice(sm.u.Object, "spec", "template", "spec", "containers")
	if err != nil {
		return "", err
	}
	if !ok || len(containers) == 0 {
		return "", fmt.Errorf("spec.template.spec.containers was missing")
	}

	container, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&containers[0])
	if err != nil {
		return "", fmt.Errorf("invalid container format")
	}

	image, ok, err := unstructured.NestedString(container, "image")
	if err != nil {
		return "", err
	}
	if !ok || image == "" {
		return "", fmt.Errorf("image was missing")
	}
	_, tag := parseContainerImage(image)

	return tag, nil
}

func parseContainerImage(image string) (name, tag string) {
	parts := strings.Split(image, ":")
	if len(parts) == 2 {
		tag = parts[1]
	}
	paths := strings.Split(parts[0], "/")
	name = paths[len(paths)-1]
	return
}
