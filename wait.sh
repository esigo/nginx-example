#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

_wait() {
until kubectl wait --for=condition=ready $* >/dev/null 2>&1; do
    echo "Waiting for '$*'"
    sleep 1
done
}

$*
