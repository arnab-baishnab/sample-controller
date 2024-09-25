#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

source "${CODEGEN_PKG}/kube_codegen.sh"

THIS_PKG="github.com/arnab-baishnab/sample-controller"

# Generate helpers
kube::codegen::gen_helpers \
    --boilerplate "${SCRIPT_ROOT}/hack/copyright.txt" \
    "${SCRIPT_ROOT}/pkg/apis"

# Generate client code
kube::codegen::gen_client \
    --with-watch \
    --output-dir "${SCRIPT_ROOT}/pkg/generated" \
    --output-pkg "${THIS_PKG}/pkg/generated" \
    --boilerplate "${SCRIPT_ROOT}/hack/copyright.txt" \
    "${SCRIPT_ROOT}/pkg/apis"

echo "here"

OUTPUT_PATH="/home/appscodepc/go/src/github.com/arnab-baishnab/sample-controller/manifests"

# Print the current working directory
echo "Current working directory is: $(pwd)"

# Print the resolved absolute path
echo "Resolved absolute path is: $OUTPUT_PATH"

# Check if the directory exists
if [[ -d "$OUTPUT_PATH" ]]; then
  echo "Directory exists."
else
  echo "Directory does not exist. Creating it now..."
  mkdir -p "$OUTPUT_PATH"
fi

# Check if controller-gen is installed
if ! command -v controller-gen &> /dev/null; then
  echo "controller-gen could not be found, please install it first."
  exit 1
fi

# Run the controller-gen command
controller-gen rbac:roleName=my-crd-controller crd \
paths="./pkg/apis/arnabbaishnab.com/v1alpha1" \
output:crd:dir="./manifests" output:stdout
