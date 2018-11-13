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

# This is a simple script to deploy kubedag in local kubernetes cluster.

set -o errexit
set -o nounset
set -o pipefail

# compile kubedag
make kube-dag

# build docker image
docker build -t kube-dag:v1 .

# deploy kubedag
kubectl create -f ./deploy/setup-kubedag.yaml
