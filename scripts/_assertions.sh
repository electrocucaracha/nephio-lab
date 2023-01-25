#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2023
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o errexit
set -o nounset
set -o pipefail
DEBUG="${DEBUG:-false}"
[[ ${DEBUG} != "true" ]] || set -o xtrace

# shellcheck source=scripts/_utils.sh
source _utils.sh

# assert_non_empty() - This assertion checks if the expected value is not empty
function assert_non_empty {
    local input=$1
    local error_msg=$2

    [[ ${DEBUG} != "true" ]] || debug "NonEmpty Assertion - value: $1"
    [[ -n $input ]] || error "$error_msg"
}

# assert_are_equal() - This assertion checks if the inputs are equal
function assert_are_equal {
    local input=$1
    local expected=$2
    local error_msg=${3:-"got $input, want $expected"}

    [[ ${DEBUG} != "true" ]] || debug "Are equal Assertion - value: $1 expected: $2"
    [[ $input == "$expected" ]] || error "$error_msg"
}

# assert_contains() - This assertion checks if the input contains another value
function assert_contains {
    local input=$1
    local expected=$2
    local error_msg=${3:-"$input doesn't contains $expected"}

    [[ ${DEBUG} != "true" ]] || debug "Contains Assertion - value: $1 expected: $2"
    [[ $input == *"$expected"* ]] || error "$error_msg"
}
