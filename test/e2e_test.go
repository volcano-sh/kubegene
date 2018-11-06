/*
Copyright 2018 The Kubegene Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package test

import (
	"flag"
	"fmt"
	"testing"
	"strings"

	"k8s.io/apimachinery/pkg/util/uuid"
	"kubegene.io/kubegene/test/e2e"
)

var config e2e.Config

const Kubegene = "kubegene"

func init() {
	flag.StringVar(&config.KubeConfig, "kubeconfig", "", "The kube config path")
	flag.StringVar(&config.KubeDagImage, "image", "", "The kubedag image")
	flag.StringVar(&config.Namespace, "namespace", "", "Namespace to run the test")
	flag.Set("logtostderr", "true")
	flag.Parse()
}

func TestKubegene(t *testing.T) {
	if len(config.Namespace) == 0 {
		config.Namespace = rand(Kubegene)
	}

	if len(config.KubeDagImage) == 0 {
		t.Fatalf("-image must be provided")
	}

	e2e.Test(t, &config)
}

func rand(prefix string) string {
	uuid := uuid.NewUUID()
	randStr := strings.Replace(string(uuid), "-", "", -1)[0:5]
	str := fmt.Sprintf("%s-%s", prefix, randStr)
	return str
}

