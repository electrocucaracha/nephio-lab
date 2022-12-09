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

function get_pkg {
    local path="$1"
    local url="$2"

    if ! [ -d "$path" ]; then
        sudo -E kpt pkg get --for-deployment "$url" "$path"
        sudo chown -R "$USER": "$path"
    fi
    if [[ ${DEBUG:-false} == "true" ]]; then
        kpt pkg tree "$path"
    fi
}

function install_pkg {
    local pkg="$1"
    local path="$base_path/$pkg"
    local url="${2:-https://github.com/nephio-project/nephio-packages.git/nephio-$pkg}"
    vm_ip=$(ip route get 8.8.8.8 | grep "^8." | awk '{ print $7 }')

    get_pkg "$path" "$url"
    # TODO: kpt fn eval "$path" --save --type mutator --image search-replace by-path=spec.git.repo by-value-regex='https://[a-zA-Z-]+:3000/(.*)' put-value="https://${vm_ip}.com/\${1}"
    find "$path" -type f -exec sed -i "s/gitea-server/$vm_ip/g" {} +
    kpt fn render "$path"
    kpt live init "$path" --force
    kpt live apply "$path" --reconcile-timeout=15m
}

sudo mkdir -p "$base_path"

for context in $(kubectl ctx); do
    kubectl ctx "$context"
    if [[ $context == "kind-nephio"* ]]; then
        install_pkg system
        install_pkg webui
        install_pkg "$participant" "https://github.com/electrocucaracha/nephio-lab.git/packages/participant"
        KUBE_EDITOR="sed -i \"s|type\: ClusterIP|type\: NodePort|g\"" kubectl -n nephio-webui edit service nephio-webui
        KUBE_EDITOR="sed -i \"s|nodePort\: .*|nodePort\: 30007|g\"" kubectl -n nephio-webui edit service nephio-webui
    else
        install_pkg "${context#*-}" https://github.com/nephio-project/nephio-packages.git/nephio-configsync
    fi
done
