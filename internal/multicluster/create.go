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
	"net"
	"os"

	"github.com/pkg/errors"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
)

func (p KindDataSource) Create(name, configPath string) error {
	clustersInfo, err := p.configReader.GetClustersInfo(configPath)
	if err != nil {
		return errors.Wrap(err, "failed to get clusters information")
	}

	// create the container to emulate the WAN network
	wanName, err := p.wanProvider.Create(name)
	if err != nil {
		return errors.Wrapf(err, "failed to create %s wan container", name)
	}

	for clusterName, clusterConfig := range *clustersInfo {
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
		gateway, err := p.createNetwork(clusterConfig.NodeSubnet, clusterName, wanName)
		if err != nil {
			return err
		}

		if err := p.createCluster(clusterName, config); err != nil {
			if err := p.deleteNetwork(clusterName); err != nil {
				return errors.Wrap(err, "failed to delete network during the cluster creation")
			}

			return errors.Wrap(err, "failed to create cluster")
		}

		if err := p.connectCluster(clusterName, wanName, gateway, svcSubnet, podSubnet); err != nil {
			return errors.Wrapf(err, "failed to connect cluster to %s network", wanName)
		}
	}

	return nil
}

func (p KindDataSource) connectCluster(clusterName, wanName, gateway, serviceSubnet,
	podSubnet string,
) error {
	// change the default network in all nodes
	// to use the wanem container and provide
	// connectivity between clusters
	nodes, err := p.clusterProvider.ListNodes(clusterName)
	if err != nil {
		return errors.Wrap(err, "failed to retrieve cluster nodes")
	}

	for _, node := range nodes {
		if err := p.dockerProvider.ReplaceGateway(node.String(), gateway); err != nil {
			return errors.Wrapf(err, "failed to replace the gateway in %s node", node)
		}
	}

	if len(nodes) == 0 {
		return nil
	}

	// insert routes in wanem to reach services through one of the nodes
	ipv4, _, err := nodes[0].IP()
	if err != nil {
		return errors.Wrapf(err, "failed to get IP address on %s", nodes[0])
	}

	if err := p.wanProvider.AddRoutes(wanName, ipv4, serviceSubnet, podSubnet); err != nil {
		return errors.Wrapf(err, "failed to add routes in %s wan emulator", wanName)
	}

	return nil
}

func (p KindDataSource) createCluster(clusterName string, config *v1alpha4.Cluster) error {
	// use the new created docker network
	os.Setenv("KIND_EXPERIMENTAL_DOCKER_NETWORK", p.getClusterNetworkName(clusterName))
	// create the cluster
	if err := p.clusterProvider.Create(
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

func (p KindDataSource) createNetwork(subnet, clusterName, wanName string) (string, error) {
	// each cluster has its own docker network with the clustername
	networkName := p.getClusterNetworkName(clusterName)
	if err := p.dockerProvider.CreateNetwork(networkName, subnet, false); err != nil {
		return "", errors.Wrapf(err, "failed to create the %s docker network", networkName)
	}

	// connect wanem with the last IP of the range
	// that the cluster will use later as gateway
	gateway, err := getLastIPSubnet(subnet)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get last IP Address from %s subnet", subnet)
	}

	if err := p.dockerProvider.ConnectNetwork(wanName, networkName, gateway.String()); err != nil {
		return "", errors.Wrapf(err, "failed to connect %s to %s network", networkName, wanName)
	}

	return gateway.String(), nil
}

func getLastIPSubnet(cidr string) (net.IP, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s CIDR", cidr)
	}

	ipAddress := ipnet.IP
	mask := ipnet.Mask

	// get the broadcast address
	lastIP := net.IP(make([]byte, len(ipAddress)))
	for i := range ipAddress {
		lastIP[i] = ipAddress[i] | ^mask[i]
	}
	// get the previous IP
	lastIP[len(ipAddress)-1]--

	return lastIP, nil
}
