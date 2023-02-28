#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2023
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o pipefail
set -o errexit
set -o nounset
[[ ${DEBUG:-false} != "true" ]] || set -o xtrace

# shellcheck source=./scripts/_assertions.sh
source _assertions.sh
# shellcheck source=scripts/_common.sh
source _common.sh

info "Assert WAN emulator image creation"
assert_non_empty "$(sudo docker images --filter reference=electrocucaracha/wanem --quiet)" "There is no WAN emulator Docker image created"

info "Assert KinD clusters creation"
assert_non_empty "$(sudo docker ps --filter label=io.x-k8s.kind.role=control-plane --quiet)" "There are no KinD clusters running"

info "Assert gitea users creation"
assert_contains "$(exec_gitea admin user list --admin)" "$gitea_admin_account"
gitea_users="$(exec_gitea admin user list)"
assert_contains "$gitea_users" cnf-vendor
assert_contains "$gitea_users" cnf-user

info "Assert gitea organization creation"
assert_contains "$(curl_gitea_api orgs)" "$nephio_gitea_org"

info "Assert gitea repos creation"
gitea_org_repos="$(curl_gitea_api "orgs/$nephio_gitea_org/repos")"
for repo in "${nephio_gitea_repos[@]}"; do
    assert_contains "$gitea_org_repos" "$repo"
done

info "Assert gitea token registration"
for context in $(kubectl config get-contexts --no-headers --output name); do
    assert_contains "$(kubectl get secrets -o jsonpath='{.items[*].metadata.name}' --context "$context")" gitea-personal-access-token
done

info "Assert MetalLB IP allocation config"
assert_contains "$(kubectl get ipaddresspools.metallb.io --namespace metallb-system -o jsonpath='{.items[*].spec.addresses}' --context kind-nephio)" '172.88.0.200-172.88.0.250'
