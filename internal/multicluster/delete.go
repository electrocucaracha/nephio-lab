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
	"strings"

	"github.com/aojea/kind-networking-plugins/pkg/docker"
	wanem "github.com/electrocucaracha/nephio-lab/internal/wan"
	"github.com/pkg/errors"
	"sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/log"
)

func Delete(clustersCfg map[string]ClusterConfig, name string) error {
	logger := kindcmd.NewLogger()
	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(logger),
	)

	wanName := wanem.GetContainerName(name)
	if err := wanem.Delete(wanName); err != nil {
		return errors.Wrapf(err, "failed to delete %s wan emulator", wanName)
	}

	clusters, err := provider.List()
	if err != nil {
		return errors.Wrap(err, "failed to retrieve the kind clusters")
	}

	for clusterName := range clustersCfg {
		for _, cluster := range clusters {
			if strings.Contains(cluster, clusterName) {
				if err = provider.Delete(cluster, ""); err != nil {
					logger.V(0).Infof("%s\n", errors.Wrapf(err, "failed to delete cluster %q", cluster))

					continue
				}

				logger.V(0).Infof("Deleted clusters: %q", cluster)
			}
		}

		if err := deleteNetwork(clusterName, logger); err != nil {
			return errors.Wrap(err, "failed to delete cluster network")
		}
	}

	return nil
}

func deleteNetwork(clusterName string, logger log.Logger) error {
	networks, err := docker.ListNetwork()
	if err != nil {
		return errors.Wrap(err, "failed to retrieve docker network list")
	}

	for _, network := range networks {
		if strings.Contains(network, clusterName) {
			if err = docker.DeleteNetwork(network); err != nil {
				logger.V(0).Infof("%s\n", errors.Wrapf(err, "failed to delete network %q", network))

				continue
			}
		}
	}

	return nil
}
