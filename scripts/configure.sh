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
# shellcheck source=./scripts/_utils.sh
source _utils.sh
# shellcheck source=./scripts/defaults.env
source defaults.env

trap get_status ERR

function _create_repo {
    curl_gitea_api "org/$1/repos" "{\"name\":\"$2\", \"auto_init\": true, \"default_branch\": \"main\"}"
}

function _create_org {
    curl_gitea_api "orgs" "{\"username\":\"$1\"}"
}

function _create_user {
    user_list=$(exec_gitea admin user list)
    if ! echo "$user_list" | grep -q "$1"; then
        user_create_cmd=(admin user create --username "$1" --password
            "$gitea_default_password" --access-token --email "$1@nephio.io")
        [[ ${2:-false} != "true" ]] || user_create_cmd+=(--admin)
        exec_gitea "${user_create_cmd[@]}"
    fi
}

function _wait_gitea_services {
    local max_attempts=5
    for svc in $(sudo docker-compose ps -aq); do
        attempt_counter=0
        until [ "$(sudo docker inspect "$svc" --format='{{.State.Health.Status}}')" == "healthy" ]; do
            [[ ${attempt_counter} -ne ${max_attempts} ]] || error "Max attempts reached for waiting to gitea containers"
            attempt_counter=$((attempt_counter + 1))
            sleep $((attempt_counter * 5))
        done
    done

    attempt_counter=0
    until curl -s http://localhost:3000/api/swagger; do
        [[ ${attempt_counter} -ne ${max_attempts} ]] || error "Max attempts reached for waiting for gitea API"
        attempt_counter=$((attempt_counter + 1))
        sleep $((attempt_counter * 5))
    done
}

# Multi-cluster configuration
if ! sudo docker ps --format "{{.Image}}" | grep -q "kindest/node"; then
    sudo multicluster create --config ./config.yml --name nephio
    mkdir -p "$HOME/.kube"
    sudo cp /root/.kube/config "$HOME/.kube/config"
    sudo chown -R "$USER": "$HOME/.kube"
    chmod 600 "$HOME/.kube/config"
fi

# Gitea configuration
if [ "${CODESPACE_NAME-}" ]; then
    gitea_domain="$CODESPACE_NAME-3000.preview.app.github.dev"
    sed -i "s|ROOT_URL .*|ROOT_URL = https://${gitea_domain}/|g" ./gitea/app.ini
fi
sudo docker-compose up -d
_wait_gitea_services

# NOTE: The first gitea user created won't be forced to change the password
_create_user "$gitea_admin_account" true
_create_user cnf-vendor
_create_user cnf-user
_create_org "$nephio_gitea_org"
for repo in "${nephio_gitea_repos[@]}"; do
    _create_repo "$nephio_gitea_org" "$repo"
done

# Wait for node readiness
for context in $(kubectl config get-contexts --no-headers --output name); do
    for node in $(kubectl get node \
        -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' \
        --context "$context"); do
        kubectl wait --for=condition=ready "node/$node" --context "$context"
        kubectl create secret generic -n default \
            gitea-personal-access-token \
            --from-literal username="$gitea_admin_account" \
            --from-literal password="$gitea_default_password" \
            --type kubernetes.io/basic-auth \
            --context "$context" || :
    done
    if [[ $context != "kind-nephio"* ]]; then
        kubectl apply --filename="https://raw.githubusercontent.com/k8snetworkplumbingwg/multus-cni/v$MULTUS_CNI_VERSION/deployments/multus-daemonset-thick-plugin.yml" --context "$context"
    else
        kubectl apply --filename="https://raw.githubusercontent.com/metallb/metallb/v$METALLB_VERSION/config/manifests/metallb-native.yaml" --context "$context"
    fi
done

for context in $(kubectl config get-contexts --no-headers --output name); do
    if [[ $context != "kind-nephio"* ]]; then
        kubectl rollout status daemonset/kube-multus-ds \
            --namespace kube-system --timeout=3m --context "$context"
    else
        kubectl wait --namespace metallb-system --for=condition=ready pod \
            --selector=app=metallb --timeout=3m --context "$context"
        kubectl apply --filename=./resources/metallb-config.yml --context "$context"
    fi
done
