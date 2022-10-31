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

if ! command -v go >/dev/null; then
    curl -fsSL http://bit.ly/install_pkg | PKG=go-lang bash
    # shellcheck disable=SC1091
    source /etc/profile.d/path.sh
fi

go mod tidy -go="$(curl -sL https://golang.org/VERSION?m=text | sed 's/go//;s/\..$//')"
GOPATH=$(go env GOPATH)
if [ ! -f "$GOPATH/bin/cyclonedx-gomod" ]; then
    go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest
fi
"$GOPATH/bin/cyclonedx-gomod" mod -licenses -json -output mod_nephio-lab.bom.json
