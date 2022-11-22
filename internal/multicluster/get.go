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

package multicluster

import (
	"fmt"
	"strings"

	wanem "github.com/electrocucaracha/nephio-lab/internal/wan"
	"github.com/pkg/errors"
	"sigs.k8s.io/kind/pkg/log"
)

type DataSource interface {
	Get(string) error
	Delete(string, string) error
	Create(string, string) error
}

type KindDataSource struct {
	configReader    ConfigReader
	wanProvider     wanem.WanProvider
	clusterProvider ClusterProvider
	dockerProvider  ContainerProvider
	logger          log.Logger
}

var _ DataSource = (*KindDataSource)(nil)

func NewProvider(configReader ConfigReader, wanProvider wanem.WanProvider,
	clusterProvider ClusterProvider, dockerProvider ContainerProvider, logger log.Logger,
) *KindDataSource {
	return &KindDataSource{
		configReader:    configReader,
		wanProvider:     wanProvider,
		clusterProvider: clusterProvider,
		dockerProvider:  dockerProvider,
		logger:          logger,
	}
}

func (p KindDataSource) getClusterNetworkName(clusterName string) string {
	return "net-" + clusterName
}

func (p KindDataSource) Get(name string) error {
	clusters, err := p.clusterProvider.List()
	if err != nil {
		return errors.Wrap(err, "failed to retrieve kind clusters")
	}

	if len(clusters) == 0 {
		p.logger.V(0).Info("No kind clusters found.")

		return nil
	}

	var inClusters []string

	clusterNamePrefix := fmt.Sprintf("multi-%s-", name)

	for _, cluster := range clusters {
		if strings.HasPrefix(cluster, clusterNamePrefix) {
			inClusters = append(inClusters, cluster)
		}
	}

	p.logger.V(0).Infof("Multicluster %s contain following clusters: %v", name, inClusters)

	return nil
}
