#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2022
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
fi

# shellcheck source=scripts/_common.sh
source _common.sh

trap get_status ERR

base_path=/opt/nephio
system_path="$base_path/system"
configsys_path="$base_path/configsync"

function get_pkg {
    local pkg="$1"
    local path="$base_path/$pkg"
    url="https://github.com/nephio-project/nephio-packages.git/nephio-$pkg"

    if ! [ -d "$path" ]; then
        sudo -E kpt pkg get --for-deployment "$url" "$path"
        sudo chown -R "$USER": "$path"
    fi
    kpt pkg tree "$path"
}

sudo mkdir -p "$base_path"

# Install server components
get_pkg system
kpt fn render "$system_path"
kpt live init "$system_path"
kubectl config use-context kind-nephio
kpt live apply "$system_path" --reconcile-timeout=15m --output=table

# Installing Config Sync in Workload Clusters
get_pkg configsync
kpt fn eval "$configsys_path" \
    --save \
    --type mutator \
    --image gcr.io/kpt-fn/search-replace:v0.2.0 \
    -- by-path=spec.git.repo by-value-regex='https://github.com/[a-zA-Z0-9-]+/(.*)' \
    put-value="https://github.com/${GITHUB_USERNAME}/\${1}"
kpt fn render "$configsys_path"
kpt live init "$configsys_path"
for i in $(seq 0 "${NUM_EDGE_CLUSTERS:-3}"); do
    kubectl config use-context "kind-edge-cluster$i"
    kpt live apply "$configsys_path" --reconcile-timeout=15m --output=table
done
