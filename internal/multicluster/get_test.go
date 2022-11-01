/*
Copyright © 2022

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package multicluster_test

import (
	"sync"

	"github.com/electrocucaracha/nephio-lab/internal/multicluster"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/kind/pkg/errors"
	"sigs.k8s.io/kind/pkg/log"
)

type mockBase struct {
	errors []error
	mutex  sync.Mutex
}

func (m *mockBase) PushErrorMessages(errorMessages []string) string {
	lastErrorMsg := ""

	for _, errMsg := range errorMessages {
		if errMsg != "" {
			lastErrorMsg = errMsg
			m.pushError(errors.New(errMsg))
		} else {
			m.pushError(nil)
		}
	}

	return lastErrorMsg
}

func (m *mockBase) pushError(err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.errors = append(m.errors, err)
}

func (m *mockBase) popError() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.errors) == 0 {
		return nil
	}

	err := m.errors[len(m.errors)-1]
	m.errors = m.errors[:len(m.errors)-1]

	return err
}

type mockClusterProvider struct {
	Clusters []string
	Nodes    []multicluster.Node
	mockBase
}

func (m *mockClusterProvider) List() ([]string, error) {
	return m.Clusters, m.popError()
}

type mockWanProvider struct {
	mockBase
}

type mockConfigReader struct {
	ClustersInfo map[string]multicluster.ClusterConfig
	mockBase
}

func (m *mockConfigReader) GetClustersInfo(string) (*map[string]multicluster.ClusterConfig, error) {
	return &m.ClustersInfo, m.popError()
}

var _ = Describe("Get Service", func() {
	var provider *multicluster.KindDataSource
	var clusterProvider *mockClusterProvider
	var wanProvider *mockWanProvider
	var containerProvider *mockContainerProvider

	BeforeEach(func() {
		logger := log.NoopLogger{}
		clusterProvider = &mockClusterProvider{}
		wanProvider = &mockWanProvider{}
		containerProvider = &mockContainerProvider{}
		provider = multicluster.NewProvider(multicluster.NewConfigReader(),
			wanProvider, clusterProvider, containerProvider, logger)
	})

	DescribeTable("get execution service process", func(clusters []string,
		clusterProviderErrorMessages []string, shouldSucceed bool,
	) {
		clusterProvider.Clusters = clusters
		errMsgExpected := clusterProvider.PushErrorMessages(clusterProviderErrorMessages)

		err := provider.Get("name")
		if shouldSucceed {
			Expect(err).NotTo(HaveOccurred())
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errMsgExpected))
		}
	},
		Entry("when no cluters exist", nil, nil, true),
		Entry("when no matching prefix clusters exists", []string{"kind"}, nil, true),
		Entry("when matching prefix clusters exists", []string{
			"multi-test-edge1",
			"multi-test-edge2", "kind",
		}, nil, true),
		Entry("when cluster provider raises a retrieval error", nil,
			[]string{"cluster provider error"}, false),
	)
})
