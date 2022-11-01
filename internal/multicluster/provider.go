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
	"github.com/pkg/errors"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/log"
)

type Node interface {
	String() string
	IP() (string, string, error)
}

type KindNode struct {
	Name string
	IPv4 string
	IPv6 string
}

func (k KindNode) String() string {
	return k.Name
}

func (k KindNode) IP() (string, string, error) {
	return k.IPv4, k.IPv6, nil
}

type ClusterProvider interface {
	List() ([]string, error)
	ListNodes(string) ([]Node, error)
	Delete(string, string) error
	Create(string, ...cluster.CreateOption) error
}

type KindProviderWrapper struct {
	provider *cluster.Provider
}

var _ ClusterProvider = (*KindProviderWrapper)(nil)

func NewClusterProvider(logger log.Logger,
) *KindProviderWrapper {
	return &KindProviderWrapper{
		provider: cluster.NewProvider(
			cluster.ProviderWithLogger(logger),
		),
	}
}

func (k *KindProviderWrapper) Create(name string, options ...cluster.CreateOption) error {
	if err := k.provider.Create(name, options...); err != nil {
		return errors.Wrap(err, "failed to create a kind node")
	}

	return nil
}

func (k *KindProviderWrapper) Delete(name, explicitKubeconfigPath string) error {
	if err := k.provider.Delete(name, explicitKubeconfigPath); err != nil {
		return errors.Wrap(err, "failed to delete a kind node")
	}

	return nil
}

func (k *KindProviderWrapper) List() ([]string, error) {
	list, err := k.provider.List()
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve the list of kind nodes")
	}

	return list, nil
}

func (k *KindProviderWrapper) ListNodes(name string) ([]Node, error) {
	nodes := []Node{}

	kindNodes, err := k.provider.ListNodes(name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve the list of kind nodes")
	}

	for _, node := range kindNodes {
		tmp := KindNode{}
		tmp.Name = node.String()
		tmp.IPv4, tmp.IPv6, _ = node.IP()

		nodes = append(nodes, tmp)
	}

	return nodes, nil
}
