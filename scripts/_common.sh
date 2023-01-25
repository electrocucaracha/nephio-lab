#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2022,2023
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o pipefail
set -o errexit
set -o nounset
[[ ${DEBUG:-false} != "true" ]] || set -o xtrace

export gitea_default_password=secret
export gitea_admin_account=gitea-admin
export nephio_gitea_org=nephio-playground
export nephio_gitea_repos=(catalog regional edge-1 edge-2)

function exec_gitea {
    sudo docker exec --user git "$(sudo docker ps --filter \
        ancestor=gitea/gitea:1.18-dev -q)" /app/gitea/gitea "$@"
}

function _get_admin_token {
    exec_gitea admin user generate-access-token --username \
        "$gitea_admin_account" | awk -F ':' '{ print $2}'
}

function curl_gitea_api {
    curl_cmd="curl -s -H 'Authorization: token $(_get_admin_token)' -H 'content-type: application/json' http://localhost:3000/api/v1/$1"
    [[ -z ${2-} ]] || curl_cmd+=" -k --data '$2'"
    eval "$curl_cmd"
}

# get_status() - Print the current status of the cluster
function get_status {
    set +o xtrace
    printf "CPU usage: "
    grep 'cpu ' /proc/stat | awk '{usage=($2+$4)*100/($2+$4+$5)} END {print usage " %"}'
    printf "Memory free(Kb):"
    awk -v low="$(grep low /proc/zoneinfo | awk '{k+=$2}END{print k}')" '{a[$1]=$2}  END{ print a["MemFree:"]+a["Active(file):"]+a["Inactive(file):"]+a["SReclaimable:"]-(12*low);}' /proc/meminfo
    echo "Kubernetes Events:"
    kubectl alpha events
    echo "Kubernetes Resources:"
    kubectl get all -A -o wide
    echo "Kubernetes Nodes:"
    kubectl describe nodes
}
