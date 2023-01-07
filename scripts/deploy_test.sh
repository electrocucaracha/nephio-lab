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
if [[ ${DEBUG:-false} == "true" ]]; then
    set -o xtrace
    export PKG_DEBUG=true
fi

# shellcheck source=./scripts/_assertions.sh
source _assertions.sh

for context in $(kubectl config get-contexts --no-headers --output name); do
    kubectl config use-context "$context"
    if [[ $context == "kind-nephio"* ]]; then
        info "Assert Nephio installation"
        nephio_system_deploy="$(kubectl get deploy -n nephio-system -o jsonpath='{.items[*].metadata.name}')"
        for deployment in ipam nephio-5gc nf-injector package-deployment-controller; do
            assert_contains "$nephio_system_deploy" "$deployment"
        done

        info "Assert Nephio UI installation"
        assert_contains "$(kubectl get deploy -n nephio-webui -o jsonpath='{.items[*].metadata.name}')" nephio-webui

        info "Assert participant"
        porch_repos="$(kubectl get repositories.config.porch.kpt.dev -o jsonpath='{.items[*].metadata.name}')"
        for repo in catalog edge-1 edge-2 free5gc-packages nephio-packages regional; do
            assert_contains "$porch_repos" "$repo"
        done
    else
        info "Assert Config Sync installation"
        assert_contains "$(kubectl get rootsyncs -n config-management-system -o jsonpath='{.items[*].metadata.name}')" nephio-workload-cluster-sync
    fi
done
