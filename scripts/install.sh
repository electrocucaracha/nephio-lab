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

# shellcheck source=./scripts/_utils.sh
source _utils.sh
# shellcheck source=./scripts/defaults.env
source defaults.env

export PKG_KREW_PLUGINS_LIST=" "

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

# Install dependencies
# NOTE: Shorten link -> https://github.com/electrocucaracha/pkg-mgr_scripts
curl -fsSL http://bit.ly/install_pkg | PKG_COMMANDS_LIST="docker,kubectl,docker-compose" PKG="go-lang cni-plugins" bash

if ! command -v kpt >/dev/null; then
    curl -s "https://i.jpillora.com/GoogleContainerTools/kpt@v${KPT_VERSION}!" | bash
    kpt completion bash | sudo tee /etc/bash_completion.d/kpt >/dev/null
fi

# shellcheck disable=SC1091
[ -f /etc/profile.d/path.sh ] && source /etc/profile.d/path.sh
for cmd in multicluster nephioadm; do
    if ! command -v "$cmd" >/dev/null; then
        GOBIN=/usr/local/bin/ sudo -E "$(command -v go)" install "github.com/electrocucaracha/$cmd/cmd/$cmd@latest"
    fi
done

# Increase inotify resources
setup_sysctl "fs.inotify.max_user_watches" "524288"
setup_sysctl "fs.inotify.max_user_instances" "512"
