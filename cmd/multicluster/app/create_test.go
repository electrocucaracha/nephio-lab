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

package app_test

import (
	"github.com/electrocucaracha/nephio-lab/cmd/multicluster/app"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

func (m mock) Create(configPath, name string) error {
	return nil
}

var _ = Describe("Create Command", func() {
	var cmd *cobra.Command

	BeforeEach(func() {
		provider := mock{}
		cmd = app.NewCreateCommand(provider)
	})

	DescribeTable("creation execution process", func(shouldSucceed bool, args ...string) {
		cmd.SetArgs(args)
		err := cmd.Execute()

		if shouldSucceed {
			Expect(err).NotTo(HaveOccurred())
		} else {
			Expect(err).To(HaveOccurred())
		}
	},
		Entry("when the default options are provided", true),
		Entry("when a multi-cluster name option is defined", true, "--name", "testName"),
		Entry("when multi-cluster name and configuration path options are defined",
			true, "--name", "testName", "--config", "config.yml"),
		Entry("when an empty multi-cluster name option is provided",
			false, "--name", ""),
		Entry("when an empty configuration path option is provided",
			false, "--name", "test", "--config", ""),
	)
})
