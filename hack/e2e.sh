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

SCRIPT_DIR=$(dirname "${BASH_SOURCE[0]}")

"${SCRIPT_DIR}"/minikube/install-minikube.sh
"${SCRIPT_DIR}"/minikube/minikube-create-cluster.sh

make docker

cd test
  make clean install || exit 1
cd ..

unset http_proxy
unset https_proxy

gene2e --image=$1 --kubeconfig=$2

rm /usr/bin/gene2e

"${SCRIPT_DIR}"/minikube/minikube-delete-cluster.sh

