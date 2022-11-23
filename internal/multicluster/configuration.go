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

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// Config struct for multicluster config.
type Config struct {
	Clusters map[string]ClusterConfig `yaml:"clusters"`
}

type ClusterConfig struct {
	*v1alpha4.Cluster
	NodeSubnet string `yaml:"nodeSubnet"`
}

type ConfigReader interface {
	GetClustersInfo(string) (*map[string]ClusterConfig, error)
}

type Reader struct{}

var _ ConfigReader = (*Reader)(nil)

func NewConfigReader() *Reader {
	return new(Reader)
}

// GetClustersInfo returns the clusters information decoded from the configuration file.
func (c Reader) GetClustersInfo(configPath string) (*map[string]ClusterConfig, error) {
	// Create config structure
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open multi-cluster configuration file")
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, errors.Wrap(err, "failed to decode multi-cluster configuration file")
	}

	return &config.Clusters, nil
}
