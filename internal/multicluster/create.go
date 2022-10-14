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
	"os"

	"github.com/aojea/kind-networking-plugins/pkg/docker"
	"github.com/aojea/kind-networking-plugins/pkg/network"
	wanem "github.com/electrocucaracha/nephio-lab/internal/wan"
	"github.com/pkg/errors"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
)

// Config struct for multicluster config.
type Config struct {
	Clusters map[string]ClusterConfig `yaml:"clusters"`
}

type ClusterConfig struct {
	Nodes         int    `yaml:"nodes"`
	NodeSubnet    string `yaml:"nodeSubnet"`
	PodSubnet     string `yaml:"podSubnet"`
	ServiceSubnet string `yaml:"serviceSubnet"`
}

func Create(clustersCfg map[string]ClusterConfig, name string) error {
	// create the container to emulate the WAN network
	wanName := wanem.GetContainerName(name)

	if err := wanem.Create(wanName); err != nil {
		return errors.Wrapf(err, "failed to create %s wan container", wanName)
	}

	logger := kindcmd.NewLogger()
	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(logger),
	)

	for clusterName, clusterConfig := range clustersCfg {
		podSubnet := clusterConfig.PodSubnet
		svcSubnet := clusterConfig.ServiceSubnet
		config := &v1alpha4.Cluster{
			Name:  clusterName,
			Nodes: createNodes(clusterConfig.Nodes),
			Networking: v1alpha4.Networking{
				PodSubnet:     podSubnet,
				ServiceSubnet: svcSubnet,
			},
		}

		// each cluster has its own docker network with the clustername
		gateway, err := createNetwork(clusterConfig.NodeSubnet, clusterName, wanName)
		if err != nil {
			return err
		}

		if err := createCluster(clusterName, config, provider); err != nil {
			if err := deleteNetwork(clusterName, logger); err != nil {
				return errors.Wrap(err, "failed to delete network during the cluster creation")
			}

			return errors.Wrap(err, "failed to create cluster")
		}

		if err := connectCluster(clusterName, wanName, gateway, svcSubnet, podSubnet, provider); err != nil {
			return errors.Wrapf(err, "failed to connect cluster to %s network", wanName)
		}
	}

	return nil
}

func connectCluster(clusterName, wanName, gateway, serviceSubnet, podSubnet string, provider *cluster.Provider) error {
	// change the default network in all nodes
	// to use the wanem container and provide
	// connectivity between clusters
	nodes, err := provider.ListNodes(clusterName)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve cluster nodes")
	}

	for _, node := range nodes {
		if err := docker.ReplaceGateway(node.String(), gateway); err != nil {
			return errors.Wrapf(err, "failed to replace the gateway in %s node", node)
		}
	}

	// insert routes in wanem to reach services through one of the nodes
	ipv4, _, err := nodes[0].IP()
	if err != nil {
		return errors.Wrapf(err, "failed to get IP address on %s", nodes[0])
	}

	if err := wanem.AddRoutes(wanName, ipv4, serviceSubnet, podSubnet); err != nil {
		return errors.Wrapf(err, "failed to add routes in %s wan emulator", wanName)
	}

	return nil
}

func createCluster(clusterName string, config *v1alpha4.Cluster, provider *cluster.Provider) error {
	// use the new created docker network
	os.Setenv("KIND_EXPERIMENTAL_DOCKER_NETWORK", clusterName)
	// create the cluster
	if err := provider.Create(
		clusterName,
		cluster.CreateWithV1Alpha4Config(config),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	); err != nil {
		return errors.Wrap(err, "failed to create kind cluster")
	}
	// reset the env variable
	os.Unsetenv("KIND_EXPERIMENTAL_DOCKER_NETWORK")

	return nil
}

func createNodes(numberNodes int) []v1alpha4.Node {
	nodes := []v1alpha4.Node{
		{
			Role: v1alpha4.ControlPlaneRole,
		},
	}

	for j := 1; j < numberNodes; j++ {
		n := v1alpha4.Node{
			Role: v1alpha4.WorkerRole,
		}
		nodes = append(nodes, n)
	}

	return nodes
}

func createNetwork(subnet, clusterName, wanName string) (string, error) {
	// each cluster has its own docker network with the clustername
	if err := docker.CreateNetwork(clusterName, subnet, false); err != nil {
		return "", errors.Wrapf(err, "failed to create the %s docker network", clusterName)
	}

	// connect wanem with the last IP of the range
	// that the cluster will use later as gateway
	gateway, err := network.GetLastIPSubnet(subnet)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get last IP Address from %s subnet", subnet)
	}

	if err := docker.ConnectNetwork(wanName, clusterName, gateway.String()); err != nil {
		return "", errors.Wrapf(err, "failed to connect %s to %s network", clusterName, wanName)
	}

	return gateway.String(), nil
}
