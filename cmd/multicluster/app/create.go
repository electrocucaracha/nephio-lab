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

package app

import (
	"os"

	"github.com/electrocucaracha/nephio-lab/internal/multicluster"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"sigs.k8s.io/kind/pkg/cluster"
)

// NewConfig returns a new decoded Config struct.
func NewConfig(configPath string) (*multicluster.Config, error) {
	// Create config structure
	config := &multicluster.Config{}

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

	return config, nil
}

func newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a deployment with multiple KIND clusters",
		Long: `Create a deployment with multiple KIND clusters based on the configuration
passed as parameters.

Multicluster deployment create KIND clusters in independent bridges, that are connected
through an special container that handles the routing and the WAN emulation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return ErrGetName
			}
			configPath, err := cmd.Flags().GetString("config")
			if err != nil {
				return ErrGetConfig
			}
			cfg, err := NewConfig(configPath)
			if err != nil {
				return err
			}

			if err := multicluster.Create(cfg.Clusters, name); err != nil {
				return errors.Wrapf(err, "failed to create %s multi-cluster", name)
			}

			return nil
		},
	}

	cmd.Flags().String(
		"name",
		cluster.DefaultName,
		"the multicluster context name",
	)

	cmd.Flags().String(
		"config",
		"./config.yml",
		"the config file with the cluster configuration",
	)

	return cmd
}
