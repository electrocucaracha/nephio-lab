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

if [ -z "$(sudo docker images wanem:0.0.1 -q)" ]; then
    sudo docker build -t wanem:0.0.1 .
fi
if [ ! -f "$HOME/go/bin/multicluster" ]; then
    pushd ../ >/dev/null
    go install ./...
    popd >/dev/null
fi
sudo -E "$HOME/go/bin/multicluster" create --config ./config.yml --name nephio
mkdir -p "$HOME/.kube"
sudo chown -R "$USER" "$HOME/.kube/"

# Wait for node readiness
for context in $(kubectl config get-contexts --no-headers --output name); do
    for node in $(kubectl get node -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' --context "$context"); do
        kubectl wait --for=condition=ready "node/$node" --timeout=3m --context "$context"
    done
done
