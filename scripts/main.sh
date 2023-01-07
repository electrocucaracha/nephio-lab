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
set -o xtrace

# shellcheck source=./scripts/_utils.sh
source _utils.sh

export DEBUG=true

for step in install configure deploy; do
    info "Running $step process"
    bash "./$step.sh"
    if [ "${ENABLE_FUNC_TEST:-false}" == "true" ]; then
        bash "./${step}_test.sh"
    fi
done

kubectl config use-context kind-nephio

if [ "${CODESPACE_NAME-}" ]; then
    if ! command -v gh >/dev/null; then
        curl -s 'https://i.jpillora.com/cli/cli!?as=gh' | bash
    fi
    for port in 7007 3000; do
        gh codespace ports visibility "$port:public" -c "$CODESPACE_NAME"
    done
    gh codespace ports -c "$CODESPACE_NAME"
fi
