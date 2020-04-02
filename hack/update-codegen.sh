#!/usr/bin/env bash

# Copyright 2018 The Kubegene Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

KUBEGENE_ROOT=$(dirname ${BASH_SOURCE})/..

# generate the code
${KUBEGENE_ROOT}/hack/generate-groups.sh "deepcopy,client,informer,lister" \
  kubegene.io/kubegene/pkg/client kubegene.io/kubegene/pkg/apis \
  gene:v1alpha1 \
  --output-base "$(dirname ${BASH_SOURCE})/../../.." \
  --go-header-file ${KUBEGENE_ROOT}/hack/boilerplate.go.txt
