# Nephio Lab
<!-- markdown-link-check-disable-next-line -->
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GitHub Super-Linter](https://github.com/electrocucaracha/nephio-lab/workflows/Lint%20Code%20Base/badge.svg)](https://github.com/marketplace/actions/super-linter)
[![Ruby Style Guide](https://img.shields.io/badge/code_style-rubocop-brightgreen.svg)](https://github.com/rubocop/rubocop)
[![Go Report Card](https://goreportcard.com/badge/github.com/electrocucaracha/nephio-lab)](https://goreportcard.com/report/github.com/electrocucaracha/nephio-lab)
[![GoDoc](https://godoc.org/github.com/electrocucaracha/nephio-lab?status.svg)](https://godoc.org/github.com/electrocucaracha/nephio-lab)
![visitors](https://visitor-badge.glitch.me/badge?page_id=electrocucaracha.nephio-lab)

The goal of this project was to automate the provision a [Nephio][1] Testing
environment through [Vagrant tool][2]. The bash provision scripts consume the
[KinD tool][3] to provision several Kubernetes clusters locally.

> This initial approach pretends to evolve as well as the Nephio project.


## Provisioning process

> It's highly recommended to use the  *setup.sh* script of the
[bootstrap-vagrant project][4] for installing Vagrant dependencies and plugins
required for this project. That script supports two Virtualization providers
(Libvirt and VirtualBox) which are determine by the **PROVIDER** environment
variable.

Once Vagrant is installed, it's possible to provision a Virtual
Machine using the following instructions:

    vagrant up

The provisioning process will take some time to install all
dependencies required by this project.

[1]: https://nephio.org/
[2]: https://www.vagrantup.com/
[3]: https://kind.sigs.k8s.io/
[4]: https://github.com/electrocucaracha/bootstrap-vagrant
