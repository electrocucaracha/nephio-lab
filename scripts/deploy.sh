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
[[ ${DEBUG:-false} != "true" ]] || set -o xtrace

# shellcheck source=scripts/_common.sh
source _common.sh

trap get_status ERR

base_path=/opt/nephio
gitea_internal_url="http://$(ip route get 8.8.8.8 | grep "^8." | awk '{ print $7 }'):3000/"
backend_base_url="http://localhost:7007"
[[ -z ${CODESPACE_NAME-} ]] || backend_base_url="https://$CODESPACE_NAME-7007.preview.app.github.dev"

function install_participant {
    local context=$1
    local path="$base_path/nephio-workshop"
    local url="https://github.com/electrocucaracha/nephio-lab.git/packages/participant"

    if ! [ -d "$path" ]; then
        sudo -E kpt pkg get --for-deployment "$url" "$path"
        sudo chown -R "$USER": "$path"
    fi
    [[ ${DEBUG:-false} != "true" ]] || kpt pkg tree "$path"
    sudo kpt fn eval "$path" --save --type mutator \
        --image gcr.io/kpt-fn/search-replace:v0.2 -- 'by-path=spec.git.repo' 'by-value-regex=http://gitea-server:3000/(.*)/(.*)' "put-value=${gitea_internal_url}${nephio_gitea_org}/\${2}"
    sudo chown -R "$USER:" ~/.kpt
    sudo kpt fn render "$path"
    sudo chown -R "$USER:" ~/.kpt
    [[ ${DEBUG:-false} != "true" ]] || kpt pkg diff "$path"
    kpt live init "$path" --force --context "$context"
    kpt live apply "$path" --reconcile-timeout=15m --context "$context"
    [[ ${DEBUG:-false} != "true" ]] || kpt live status "$path" --context "$context"
}

for context in $(kubectl config get-contexts --no-headers --output name); do
    sudo kubectl config use-context "$context"
    if [[ $context == "kind-nephio"* ]]; then
        sudo mkdir -p "$base_path/mgmt"
        sudo nephioadm init --base-path "$base_path/mgmt" --git-service "$gitea_internal_url${nephio_gitea_org}" --backend-base-url "$backend_base_url" --webui-cluster-type "${NEPHIO_WEBUI_CLUSTER_TYPE:-NodePort}"
        install_participant "$context"
    else
        sudo mkdir -p "$base_path/${context#*-}"
        sudo nephioadm join --base-path "$base_path/${context#*-}" --git-service "$gitea_internal_url${nephio_gitea_org}"
    fi
done
