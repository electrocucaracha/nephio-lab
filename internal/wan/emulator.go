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

package wanem

import (
	"github.com/pkg/errors"
	"sigs.k8s.io/kind/pkg/exec"
)

const dockerWanImage = "wanem:0.0.1"

type WanProvider interface {
	Create(string) (string, error)
	AddRoutes(string, string, ...string) error
	Delete(string) error
}

type Provider struct{}

var _ WanProvider = (*Provider)(nil)

func NewProvider() *Provider {
	return new(Provider)
}

func getContainerName(name string) string {
	return "wan-" + name
}

func (p Provider) Create(name string) (string, error) {
	containerName := getContainerName(name)
	args := []string{
		"run",
		"-d", // run in the background
		"--sysctl=net.ipv4.ip_forward=1",
		"--sysctl=net.ipv4.conf.all.rp_filter=0",
		"--privileged",
		"--name", containerName, // well known name
		dockerWanImage,
	}

	if err := exec.Command("docker", args...).Run(); err != nil {
		return "", errors.Wrapf(err, "failed to create %s wan emulator", containerName)
	}
	// configure masquerading so clusters can reach internet
	args = []string{
		"exec", containerName,
		"iptables", "-t", "nat", "-A", "POSTROUTING", "-o", "eth0", "-j", "MASQUERADE",
	}

	if err := exec.Command("docker", args...).Run(); err != nil {
		return "", errors.Wrapf(err, "failed to configure masquerading in %s wan emulator ", containerName)
	}

	return containerName, nil
}

func (p Provider) AddRoutes(containerName, gateway string, subnets ...string) error {
	for _, subnet := range subnets {
		args := []string{
			"exec", containerName,
			"ip", "route", "add", subnet, "via", gateway,
		}

		if err := exec.Command("docker", args...).Run(); err != nil {
			return errors.Wrapf(err, "failed to add docker routes in %s wan emulator", containerName)
		}
	}

	return nil
}

func (p Provider) Delete(name string) error {
	containerName := getContainerName(name)
	if err := exec.Command("docker", "rm", "-f", containerName).Run(); err != nil {
		return errors.Wrapf(err, "failed to delete %s wan emulator", containerName)
	}

	return nil
}
