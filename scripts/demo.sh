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

kubectl config use-context kind-nephio

# Registering Repositories
kpt alpha repo register \
    --namespace default \
    --repo-basic-username="${GITHUB_USERNAME}" \
    --repo-basic-password="${GITHUB_TOKEN}" \
    "https://github.com/${GITHUB_USERNAME}/nephio-test-catalog-01.git"

# Registering Deployment repository
kpt alpha repo register \
    --deployment \
    --namespace default \
    --repo-basic-username="${GITHUB_USERNAME}" \
    --repo-basic-password="${GITHUB_TOKEN}" \
    "https://github.com/${GITHUB_USERNAME}/nephio-edge-cluster-01.git"

# Register Blueprint repository
kpt alpha repo register \
    --namespace default \
    https://github.com/nephio-project/nephio-packages.git

kpt alpha repo get
