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

kube_version=$(curl -sL https://registry.hub.docker.com/v2/repositories/kindest/node/tags | python -c 'import json,sys,re;versions=[obj["name"][1:] for obj in json.load(sys.stdin)["results"] if re.match("^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$",obj["name"])];print("\n".join(versions))' | uniq | sort -rn | head -n 1)

function deploy_k8s_cluster {
    local name="$1"

    if ! kind get clusters | grep -q "$name"; then
        kind create cluster --name "$name" --image "kindest/node:v$kube_version"
    fi
}

# Create Nephio cluster
deploy_k8s_cluster nephio
# Creating workload clusters
for i in $(seq 0 "${NUM_EDGE_CLUSTERS:-3}"); do
    deploy_k8s_cluster "edge-cluster$i"
done

# Wait for node readiness
for context in $(kubectl config get-contexts --no-headers --output name); do
    kubectl config use-context "$context"
    for node in $(kubectl get node -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}'); do
        kubectl wait --for=condition=ready "node/$node" --timeout=3m
    done
done
