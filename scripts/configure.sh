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

function exec_gitea {
    sudo docker exec --user git "$(sudo docker ps --filter \
        ancestor=gitea/gitea:1.18-dev -q)" /app/gitea/gitea "$@"
}

function get_admin_token {
    exec_gitea admin user generate-access-token --username \
        "$gitea_admin_account" | awk -F ':' '{ print $2}'
}

function _curl_gitea_api {
    curl -k --data "$2" \
        -H "Authorization: token $(get_admin_token)" \
        -H "content-type: application/json" \
        "http://localhost:3000/api/v1/$1"
}

function create_repo {
    _curl_gitea_api "org/$1/repos" "{\"name\":\"$2\", \"auto_init\": true, \"default_branch\": \"main\"}"
}

function create_org {
    _curl_gitea_api "orgs" "{\"username\":\"$1\"}"
}

function create_user {
    user_list=$(exec_gitea admin user list)
    if ! echo "$user_list" | grep -q "$1"; then
        user_create_cmd=(admin user create --username "$1" --password
            "$gitea_default_password" --access-token --email "$1@nephio.io")
        if [ "${2:-false}" == "true" ]; then
            user_create_cmd+=(--admin)
        fi
        exec_gitea "${user_create_cmd[@]}"
    fi
}

# Multi-cluster configuration
if [ -z "$(sudo docker images wanem:0.0.1 -q)" ]; then
    sudo docker build -t wanem:0.0.1 .
fi
if [ ! -f "$HOME/go/bin/multicluster" ]; then
    pushd ../ >/dev/null
    go install ./...
    popd >/dev/null
fi
if ! sudo docker ps --format "{{.Image}}" | grep -q "kindest/node"; then
    sudo -E "$HOME/go/bin/multicluster" create --config ./config.yml --name nephio
    mkdir -p "$HOME/.kube"
    sudo chown -R "$USER" "$HOME/.kube/"
fi

# Gitea configuration
sudo docker-compose up -d
attempt_counter=0
max_attempts=5
while ! sudo docker-compose logs frontend | grep -q "Starting new Web server"; do
    if [ ${attempt_counter} -eq ${max_attempts} ]; then
        echo "Max attempts reached"
        exit 1
    fi
    attempt_counter=$((attempt_counter + 1))
    sleep $((attempt_counter * 5))
done

# NOTE: The first gitea user created won't be forced to change the password
create_user "$gitea_admin_account" true
create_user cnf-vendor
create_user cnf-user
create_org "$nephio_gitea_org"
for repo in catalog regional edge-1 edge-2; do
    create_repo "$nephio_gitea_org" "$participant-$repo"
done

# Wait for node readiness
for context in $(kubectl ctx); do
    for node in $(kubectl get node \
        -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' \
        --context "$context"); do
        kubectl wait --for=condition=ready "node/$node" --context "$context"
        kubectl create secret generic -n default \
            gitea-personal-access-token \
            --from-literal username="$gitea_admin_account" \
            --from-literal password="$(get_admin_token)" \
            --type kubernetes.io/basic-auth \
            --context "$context"
    done
done

kubectl create secret generic -n default \
    gitea-personal-access-token \
    --from-literal username="$gitea_admin_account" \
    --from-literal password="$gitea_default_password" \
    --type kubernetes.io/basic-auth \
    --context kind-nephio
