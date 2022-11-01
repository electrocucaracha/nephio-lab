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
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

type ContainerProvider interface {
	CreateNetwork(string, string, bool) error
	ConnectNetwork(string, string, string) error
	ReplaceGateway(string, string) error
	ListNetwork() ([]string, error)
	DeleteNetwork(name string) error
}

const (
	subnetIPRangeOnes = 27
	subnetIPRangeBits = 32
)

type DockerProvider struct {
	ExecCommand func(string, ...string) *exec.Cmd
}

var _ ContainerProvider = (*DockerProvider)(nil)

func NewDockerProvider() *DockerProvider {
	return &DockerProvider{
		ExecCommand: exec.Command,
	}
}

func (d *DockerProvider) ConnectNetwork(nameOrID, network, ip string) error {
	args := []string{"network", "connect"}
	if ip != "" {
		args = append(args, "--ip", ip)
	}

	args = append(args, network, nameOrID)

	if err := d.ExecCommand("docker", args...).Run(); err != nil {
		return errors.Wrapf(err, "failed to connect %s container to %s network", nameOrID, network)
	}

	return nil
}

func (d *DockerProvider) ReplaceGateway(name, gw string) error {
	gateway := net.ParseIP(gw)
	if gateway.To4() == nil {
		return ErrUnsupportedIP
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	pid, err := d.getContainerPid(name)
	if err != nil {
		return err
	}

	namespace, err := netns.GetFromPid(pid)
	if err != nil {
		return errors.Wrapf(err, "failed to get namespace from %d process id", pid)
	}
	defer namespace.Close()

	// Switch to the docker namespace to get the container interfaces
	if err := netns.Set(namespace); err != nil {
		return errors.Wrapf(err, "failed to switch to %s namespace", namespace)
	}

	defaultRoute := &netlink.Route{
		Dst: nil,
		Gw:  gateway,
	}

	if err := netlink.RouteReplace(defaultRoute); err != nil {
		return errors.Wrap(err, "failed to replace default routes")
	}

	return nil
}

// CreateNetwork create a docker network with the passed parameters.
func (d *DockerProvider) CreateNetwork(name, subnet string, masquerade bool) error {
	args := []string{"network", "create", "-d=bridge"}
	// enable docker iptables rules to masquerade network traffic
	args = append(args, "-o", fmt.Sprintf("com.docker.network.bridge.enable_ip_masquerade=%t", masquerade))
	// configure the subnet and the gateway provided
	if subnet != "" {
		args = append(args, "--subnet", subnet)
		// and only allocate ips for the containers for the first 32 ips /27
		_, cidr, err := net.ParseCIDR(subnet)
		if err != nil {
			return errors.Wrap(err, "failed to parsed the subnet CIDR")
		}

		m := net.CIDRMask(subnetIPRangeOnes, subnetIPRangeBits)
		cidr.Mask = m
		args = append(args, "--ip-range", cidr.String())
	}

	args = append(args, name)

	if err := d.ExecCommand("docker", args...).Run(); err != nil {
		return errors.Wrapf(err, "failed to create %s docker network", name)
	}

	return nil
}

// DeleteNetwork delete a docker network.
func (d *DockerProvider) DeleteNetwork(name string) error {
	if err := d.ExecCommand("docker", "network", "rm", name).Run(); err != nil {
		return errors.Wrapf(err, "failed to delete %s docker network", name)
	}

	return nil
}

func (d *DockerProvider) getContainerPid(name string) (int, error) {
	cmd := d.ExecCommand("docker", "inspect",
		"--format", `{{ .State.Pid }}`, name)
	lines, err := OutputLines(cmd)

	if err != nil || len(lines) != 1 {
		return 0, errors.Wrapf(err, "error trying to get container %s id", name)
	}

	pid, err := strconv.Atoi(lines[0])
	if err != nil {
		return 0, errors.Wrap(err, "failed to get container process ID")
	}

	return pid, nil
}

func (d *DockerProvider) ListNetwork() ([]string, error) {
	cmd := d.ExecCommand("docker", "network", "list",
		"--format", `{{ .Name }}`)

	return OutputLines(cmd)
}

func OutputLines(cmd *exec.Cmd) ([]string, error) {
	var buff bytes.Buffer
	cmd.Stdout = &buff

	if err := cmd.Run(); err != nil {
		return nil, errors.Wrap(err, "failed to run command")
	}

	scanner := bufio.NewScanner(&buff)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, nil
}
