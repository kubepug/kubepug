#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

if [[ ! -f kubepugImagerefs ]]; then
    echo "kubepugImagerefs not found"
    exit 1
fi

echo "Signing images with Keyless..."
readarray -t kubepug_images < <(cat kubepugImagerefs || true)
cosign sign --yes "${kubepug_images[@]}"
cosign verify --certificate-identity-regexp ".*" --certificate-oidc-issuer-regexp ".*" "${kubepug_images[@]}"
