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

package multicluster_test

import (
	"os"
	"os/exec"
	"strconv"
	"testing"

	"github.com/electrocucaracha/nephio-lab/internal/multicluster"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func fakeExecCommand(exitCode int) func(string, ...string) *exec.Cmd {
	return func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestDockerCommand", "--", command}
		cs = append(cs, args...)
		osCommand := os.Args[0]
		cmd := exec.Command(osCommand, cs...)
		code := strconv.Itoa(exitCode)
		cmd.Env = []string{
			"GO_WANT_HELPER_PROCESS=1",
			"EXIT_CODE=" + code,
		}

		return cmd
	}
}

func TestDockerCommand(t *testing.T) {
	t.Parallel()

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	exitCode := 0

	if val, ok := os.LookupEnv("EXIT_CODE"); ok {
		if code, err := strconv.ParseInt(val, 10, 0); err == nil {
			exitCode = int(code)
		}
	}

	os.Exit(exitCode)
}

var _ = Describe("Docker Service", func() {
	var provider *multicluster.DockerProvider

	BeforeEach(func() {
		provider = multicluster.NewDockerProvider()
	})

	DescribeTable("Connect network process", func(nameOrID, network, ipAddress string,
		exitCode int, shouldSucceed bool,
	) {
		provider.ExecCommand = fakeExecCommand(exitCode)
		defer func() { provider.ExecCommand = exec.Command }()

		err := provider.ConnectNetwork(nameOrID, network, ipAddress)
		if shouldSucceed {
			Expect(err).NotTo(HaveOccurred())
		} else {
			Expect(err).To(HaveOccurred())
		}
	},
		Entry("valid container name, network and IP Address",
			"testName", "testNetwork", "0.0.0.0", 0, true),
		Entry("valid container name, network and IP Address but docker command issues",
			"testName", "testNetwork", "0.0.0.0", 1, false),
	)

	DescribeTable("Create a docker network", func(name, network string, masquerade bool,
		exitCode int, shouldSucceed bool,
	) {
		provider.ExecCommand = fakeExecCommand(exitCode)
		defer func() { provider.ExecCommand = exec.Command }()

		err := provider.CreateNetwork(name, network, masquerade)
		if shouldSucceed {
			Expect(err).NotTo(HaveOccurred())
		} else {
			Expect(err).To(HaveOccurred())
		}
	},
		Entry("valid network name, empty subnet and masquerade enabled",
			"testName", "", true, 0, true),
		Entry("valid network name and subnet, and masquerade enabled",
			"testName", "0.0.0.0/24", true, 0, true),
		Entry("valid network name and subnet, and masquerade enabled but docker command issues",
			"testName", "0.0.0.0/24", true, 1, false),
		Entry("valid network name and invalid subnet, and masquerade enabled",
			"testName", "invalid", true, 0, false),
	)
})
