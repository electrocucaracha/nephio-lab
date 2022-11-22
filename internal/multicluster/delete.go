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

	"github.com/pkg/errors"
)

func (p KindDataSource) Delete(name, configPath string) error {
	clustersInfo, err := p.configReader.GetClustersInfo(configPath)
	if err != nil {
		return errors.Wrap(err, "failed to get clusters information")
	}

	if err := p.wanProvider.Delete(name); err != nil {
		return errors.Wrapf(err, "failed to delete %s wan emulator", name)
	}

	clusters, err := p.clusterProvider.List()
	if err != nil {
		return errors.Wrap(err, "failed to retrieve the kind clusters")
	}

	for clusterName := range *clustersInfo {
		for _, cluster := range clusters {
			if strings.Contains(cluster, clusterName) {
				if err = p.clusterProvider.Delete(cluster, ""); err != nil {
					p.logger.V(0).Infof("%s\n", errors.Wrapf(err, "failed to delete cluster %q", cluster))

					continue
				}

				p.logger.V(0).Infof("Deleted clusters: %q", cluster)
			}
		}

		if err := p.deleteNetwork(p.getClusterNetworkName(clusterName)); err != nil {
			return errors.Wrap(err, "failed to delete cluster network")
		}
	}

	return nil
}

func (p KindDataSource) deleteNetwork(clusterName string) error {
	networks, err := p.dockerProvider.ListNetwork()
	if err != nil {
		return errors.Wrap(err, "failed to retrieve docker network list")
	}

	for _, network := range networks {
		if strings.Contains(network, clusterName) {
			if err = p.dockerProvider.DeleteNetwork(network); err != nil {
				p.logger.V(0).Infof("%s\n", errors.Wrapf(err, "failed to delete network %q", network))

				continue
			}
		}
	}

	return nil
}
