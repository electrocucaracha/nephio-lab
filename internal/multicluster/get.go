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

	"github.com/pkg/errors"
	"sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
)

func Get(name string) error {
	logger := kindcmd.NewLogger()

	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(logger),
	)

	clusters, err := provider.List()
	if err != nil {
		return errors.Wrap(err, "failed to retrieve kind clusters")
	}

	if len(clusters) == 0 {
		logger.V(0).Info("No kind clusters found.")

		return nil
	}

	var inClusters []string

	clusterNamePrefix := fmt.Sprintf("multi-%s-", name)

	for _, cluster := range clusters {
		if strings.Contains(cluster, clusterNamePrefix) {
			inClusters = append(inClusters, cluster)
		}
	}

	logger.V(0).Infof("Multicluster %s contain following clusters: %v", name, inClusters)

	return nil
}
