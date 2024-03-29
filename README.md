# Nephio Lab
<!-- markdown-link-check-disable-next-line -->
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GitHub Super-Linter](https://github.com/electrocucaracha/nephio-lab/workflows/Lint%20Code%20Base/badge.svg)](https://github.com/marketplace/actions/super-linter)
[![Ruby Style Guide](https://img.shields.io/badge/code_style-rubocop-brightgreen.svg)](https://github.com/rubocop/rubocop)
<!-- markdown-link-check-disable-next-line -->
![visitors](https://visitor-badge.laobi.icu/badge?page_id=electrocucaracha.nephio-lab)

The goal of this project is to provision a self-contained environment for the
[Nephio][1] [Workshop][2] hosted during the ONE Summit 2022. The
[post create command script](./scripts/main.sh) installs
dependencies and configures Nephio services, it uses the [Multi-Cluster tool][3]
to deploy several Kubernetes clusters locally and [Gitea][4] as software package
hosting service. Nephio UI consumes [Backstage][5] project and local ports are
forwarded as follows:

<!-- markdown-link-check-disable -->
* Software Package hosting URL - <http://localhost:3000/>
* Nephio UI URL - <http://localhost:7007/>
<!-- markdown-link-check-enable -->

> This initial approach pretends to evolve as well as the Nephio project.

## Provisioning process

This project supports two provisioning methods:

* Local (via [Vagrant tool][6])
> It's highly recommended to use the  *setup.sh* script of the
[bootstrap-vagrant project][7] for installing Vagrant dependencies and plugins
required for this project. That script supports two Virtualization providers
(Libvirt and VirtualBox) which are determine by the **PROVIDER** environment
variable.

* Remote (via [Codespaces][8])
[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://github.com/codespaces/new?hide_repo_select=true&ref=master&repo=538643510)
> [Apparently][9] the only way to access to tcpservers (like Nephio UI) in
Codespaces is [using the VSCode client][10], which binds all ports to
`127.0.0.1` address.

The following diagram shows the result after its execution.

```text
+---------------------------------+     +---------------------------------+     +---------------------------------+     +---------------------------------+
| nephio (k8s)                    |     | regional (k8s)                  |     | edge-1 (k8s)                    |     | edge-2 (k8s)                    |
| +-----------------------------+ |     | +-----------------------------+ |     | +-----------------------------+ |     | +-----------------------------+ |
| | nephio-control-plane        | |     | | regional-control-plane      | |     | | edge-1-control-plane        | |     | | edge-2-control-plane        | |
| | podSubnet: 10.196.0.0/16    | |     | | podSubnet: 10.197.0.0/16    | |     | | podSubnet: 10.198.0.0/16    | |     | | podSubnet: 10.199.0.0/16    | |
| | serviceSubnet: 10.96.0.0/16 | |     | | serviceSubnet: 10.97.0.0/16 | |     | | serviceSubnet: 10.98.0.0/16 | |     | | serviceSubnet: 10.99.0.0/16 | |
| +-----------------------------+ |     | +-----------------------------+ |     | +-----------------------------+ |     | +-----------------------------+ |
| | eth0(172.88.0.2/16)         | |     | | eth0(172.89.0.2/16)         | |     | | eth0(172.90.0.2/16)         | |     | | eth0(172.91.0.2/16)         | |
| +------------+----------------+ |     | +------------+----------------+ |     | +------------+----------------+ |     | +------------+----------------+ |
|              |                  |     |              |                  |     |              |                  |     |              |                  |
+--------------+------------------+     +--------------+------------------+     +--------------+------------------+     +--------------+------------------+
               |                                       |                                       |                                       |
     +=========+============+                +=========+============+                +=========+===========+                 +=========+===========+
     |  net-nephio(bridge)  |                | net-regional(bridge) |                |  net-edge-1(bridge) |                 |  net-edge-2(bridge) |
     |    172.88.0.0/16     |                |    172.89.0.0/16     |                |    172.90.0.0/16    |                 |    172.91.0.0/16    |
     +=========+============+                +=========+============+                +=========+===========+                 +=========+===========+
               |                                       |                                       |                                       |
+--------------+---------------------------------------+---------------------------------------+---------------------------------------+-----------+
| wan-nephio (emulator)                                                                                                                            |
+--------------------------------------------------------------------------------------------------------------------------------------------------+
| eth0(172.80.0.2/24)                                                                                                                              |
| eth1(172.90.0.254/16)                                                                                                                            |
| eth2(172.91.0.254/16)                                                                                                                            |
| eth3(172.89.0.254/16)                                                                                                                            |
| eth4(172.88.0.254/16)                                                                                                                            |
+--------------------------------------------------------------------------------------------------------------------------------------------------+

+===================================+
|          host(host)               |
+========+===================+======+
         |                   |
+--------+---------+ +-------+------+
| frontend (gitea) | | db (mariadb) |
+------------------+ +--------------+
|                  | |              |
+------------------+ +--------------+
```

[1]: https://nephio.org/
[2]: https://github.com/nephio-project/one-summit-22-workshop/
[3]: https://github.com/electrocucaracha/multi-cluster
[4]: https://gitea.io/
[5]: https://backstage.io/
[6]: https://www.vagrantup.com/
[7]: https://github.com/electrocucaracha/bootstrap-vagrant
[8]: https://github.com/features/codespaces
[9]: https://github.com/orgs/community/discussions/28528#discussioncomment-3360324
[10]: https://docs.github.com/en/codespaces/developing-in-codespaces/using-github-codespaces-in-visual-studio-code
