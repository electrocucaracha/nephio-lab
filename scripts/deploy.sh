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
nephio_url_base="https://github.com/nephio-project/nephio-packages.git/nephio-"
gitea_internal_url="http://$(ip route get 8.8.8.8 | grep "^8." | awk '{ print $7 }'):3000/"

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

function install_system {
    local path="$base_path/system"

    get_pkg "$path" "${nephio_url_base}system"
    _install_pkg "$path"
}

function install_webui {
    local path="$base_path/webui"
    local nephio_webui_cluster_type=${NEPHIO_WEBUI_CLUSTER_TYPE:-NodePort}

    get_pkg "$path" "${nephio_url_base}webui"
    if [ "${CODESPACE_NAME-}" ]; then
        sed -i "s|baseUrl: .*|baseUrl: https://$CODESPACE_NAME-7007.preview.app.github.dev|g" "${path}/config-map.yaml"
    fi
    _install_pkg "$path"
    KUBE_EDITOR="sed -i \"s|type\: .*|type\: $nephio_webui_cluster_type|g\"" kubectl -n nephio-webui edit service nephio-webui
    [ "$nephio_webui_cluster_type" != "NodePort" ] || KUBE_EDITOR="sed -i \"s|nodePort\: .*|nodePort\: 30007|g\"" kubectl -n nephio-webui edit service nephio-webui
}

function install_configsync {
    local path="$base_path/$1"

    get_pkg "$path" "${nephio_url_base}configsync"
    kpt fn eval "$path" --save --type mutator \
        --image gcr.io/kpt-fn/search-replace:v0.2 -- 'by-path=spec.git.repo' 'by-value-regex=https://github.com/(.*)/(.*)' "put-value=${gitea_internal_url}${nephio_gitea_org}/\${2}"
    _install_pkg "$path"
}

function install_participant {
    local path="$base_path/nephio-workshop"

    get_pkg "$path" https://github.com/electrocucaracha/nephio-lab.git/packages/participant
    kpt fn eval "$path" --save --type mutator \
        --image gcr.io/kpt-fn/search-replace:v0.2 -- 'by-path=spec.git.repo' 'by-value-regex=http://gitea-server:3000/(.*)/(.*)' "put-value=${gitea_internal_url}${nephio_gitea_org}/\${2}"
    _install_pkg "$path"
}

function _install_pkg {
    local path="$1"

    kpt fn render "$path"
    kpt live init "$path" --force
    kpt live apply "$path" --reconcile-timeout=15m
}

sudo mkdir -p "$base_path"

for context in $(kubectl config get-contexts --no-headers --output name); do
    kubectl config use-context "$context"
    if [[ $context == "kind-nephio"* ]]; then
        install_system
        install_webui
        install_participant
    else
        install_configsync "${context#*-}"
    fi
done
