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
	wanem "github.com/electrocucaracha/nephio-lab/internal/wan"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
)

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multicluster",
		Short: "Simulate multicluster deployments using KIND clusters",
		Long: `Simulate multicluster deployments using KIND clusters.
	Multicluster creates KIND clusters in independent bridges, that are connected
	through an special container that handles the routing and the WAN emulation.
	`,
	}

	logger := kindcmd.NewLogger()
	provider := multicluster.NewProvider(multicluster.NewConfigReader(), wanem.NewProvider(),
		multicluster.NewClusterProvider(logger), multicluster.NewDockerProvider(), logger)

	cmd.AddCommand(NewCreateCommand(provider))
	cmd.AddCommand(NewDeleteCommand(provider))
	cmd.AddCommand(NewGetCommand(provider))

	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

func getName(flags *pflag.FlagSet) (string, error) {
	name, _ := flags.GetString("name")

	if name == "" {
		return "", ErrEmptyName
	}

	return name, nil
}

func getConfigPath(flags *pflag.FlagSet) (string, error) {
	configPath, _ := flags.GetString("config")

	if configPath == "" {
		return "", ErrEmptyConfigPath
	}

	return configPath, nil
}
