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
    export PKG_DEBUG=true
fi

export PKG_KREW_PLUGINS_LIST=" "

function get_github_latest_tag {
    version=""
    attempt_counter=0
    max_attempts=5

    until [ "$version" ]; do
        tags="$(curl -s "https://api.github.com/repos/$1/tags")"
        if [ "$tags" ]; then
            version="$(echo "$tags" | grep -Po '"name":.*?[^\\]",' | awk -F '"' 'NR==1{print $4}')"
            break
        elif [ ${attempt_counter} -eq ${max_attempts} ]; then
            echo "Max attempts reached"
            exit 1
        fi
        attempt_counter=$((attempt_counter + 1))
        sleep $((attempt_counter * 2))
    done

    echo "${version#*v}"
}

function setup_sysctl {
    local key="$1"
    local value="$2"

    if [ "$(sysctl -n "$key")" != "$value" ]; then
        if [ -d /etc/sysctl.d ]; then
            echo "$key=$value" | sudo tee "/etc/sysctl.d/99-$key.conf"
        elif [ -f /etc/sysctl.conf ]; then
            echo "$key=$value" | sudo tee --append /etc/sysctl.conf
        fi

        sudo sysctl "$key=$value"
    fi
}

if [ -f /etc/netplan/01-netcfg.yaml ]; then
    sudo sed -i "s/addresses: .*/addresses: [1.1.1.1, 8.8.8.8, 8.8.4.4]/g" /etc/netplan/01-netcfg.yaml
    sudo netplan apply
fi

# Install dependencies
# NOTE: Shorten link -> https://github.com/electrocucaracha/pkg-mgr_scripts
curl -fsSL http://bit.ly/install_pkg | PKG_COMMANDS_LIST="pip,docker,kubectl,docker-compose" PKG="go-lang cni-plugins" bash

if ! command -v kpt >/dev/null; then
    curl -s "https://i.jpillora.com/GoogleContainerTools/kpt@v$(get_github_latest_tag GoogleContainerTools/kpt)!!" | bash
    kpt completion bash | sudo tee /etc/bash_completion.d/kpt >/dev/null
fi

# Increase inotify resources
setup_sysctl "fs.inotify.max_user_watches" "524288"
setup_sysctl "fs.inotify.max_user_instances" "512"
