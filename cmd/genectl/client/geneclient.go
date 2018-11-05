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

package client

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"kubegene.io/kubegene/cmd/genectl/util"
	execclientset "kubegene.io/kubegene/pkg/client/clientset/versioned"
)

func GetGeneClient(cmd *cobra.Command) (execclientset.Interface, error) {
	clientConfig := GetkubeClientConfig(cmd)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	// create the clientset
	geneClient, err := execclientset.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return geneClient, nil
}

func GetkubeClientConfig(cmd *cobra.Command) clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// use the standard defaults for this client command
	// DEPRECATED: remove and replace with something more accurate
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	kubeConfig := util.GetFlagString(cmd, "kubeconfig")

	if kubeConfig != "" {
		loadingRules.ExplicitPath = kubeConfig
	}

	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)

	return clientConfig
}
