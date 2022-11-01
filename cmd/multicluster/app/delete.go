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
	"github.com/electrocucaracha/nephio-lab/internal/multicluster"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/cluster"
)

func NewDeleteCommand(provider multicluster.DataSource) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete the specified multicluster",
		Long:  `Delete the specified multicluster`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := getName(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "failed to retrieve the name of the multi-cluster")
			}

			configPath, err := getConfigPath(cmd.Flags())
			if err != nil {
				return errors.Wrap(err, "failed to retrieve the configuration file path of the multi-cluster")
			}

			if err := provider.Delete(name, configPath); err != nil {
				return errors.Wrap(err, "failed to delete multi-cluster")
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
