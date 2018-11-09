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

package e2e

import (
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	execclientset "kubegene.io/kubegene/pkg/client/clientset/versioned"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var gtc *GeneTestContext

// Config provides the configuration for the e2e tests.
type Config struct {
	KubeConfig   string
	KubeDagImage string
	Namespace    string
}

// GeneTestContext holds the variables that each test can depend on. It
// gets initialized before each test block runs.
type GeneTestContext struct {
	Config     *Config
	KubeClient kubernetes.Interface
	GeneClient execclientset.Interface
}

func Test(t *testing.T, config *Config) {
	gtc = &GeneTestContext{
		Config: config,
	}

	registerTestsInGinkgo(gtc)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubegene Test Suite")
}

var _ = BeforeSuite(func() {
	gtc.setup()
})

var _ = AfterSuite(func() {
	gtc.teardown()
})

func (gtc *GeneTestContext) setup() {
	// build client from kube config
	gtc.buildClient()

	// create kubedag deployment
	gtc.createKubedag()
}

func (gtc *GeneTestContext) teardown() {
	By("Delete service account")
	err := DeleteServiceAccount(
		gtc.KubeClient,
		gtc.Config.Namespace,
		"deploy/serviceAccount.yaml",
	)
	Expect(err).NotTo(HaveOccurred())

	By("Delete cluster role")
	err = DeleteClusterRole(gtc.KubeClient, "deploy/clusterRole.yaml")
	Expect(err).NotTo(HaveOccurred())

	By("Delete cluster role binding")
	err = DeleteClusterRoleBinding(gtc.KubeClient, "deploy/clusterRoleBinding.yaml")
	Expect(err).NotTo(HaveOccurred())

	By("Delete deployment")
	err = DeleteDeployment(
		gtc.KubeClient,
		gtc.Config.Namespace,
		"deploy/deployment.yaml")

	Expect(err).NotTo(HaveOccurred())

	err = gtc.KubeClient.CoreV1().Namespaces().Delete(gtc.Config.Namespace, &metav1.DeleteOptions{})
	Expect(err).NotTo(HaveOccurred())
}

func (gtc *GeneTestContext) buildClient() {
	config, err := clientcmd.BuildConfigFromFlags("", gtc.Config.KubeConfig)
	Expect(err).NotTo(HaveOccurred())

	kubeClient, err := kubernetes.NewForConfig(config)
	Expect(err).NotTo(HaveOccurred())
	gtc.KubeClient = kubeClient

	geneClient, err := execclientset.NewForConfig(config)
	Expect(err).NotTo(HaveOccurred())
	gtc.GeneClient = geneClient
}

func (gtc *GeneTestContext) createKubedag() {
	_, err := gtc.KubeClient.CoreV1().Namespaces().Get(gtc.Config.Namespace, metav1.GetOptions{})
	if err != nil {
		namespace := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: gtc.Config.Namespace,
			},
		}
		_, err := gtc.KubeClient.CoreV1().Namespaces().Create(namespace)
		Expect(err).NotTo(HaveOccurred())
	}

	By("Create service account")
	err = CreateServiceAccount(
		gtc.KubeClient,
		gtc.Config.Namespace,
		"deploy/serviceAccount.yaml",
	)
	Expect(err).NotTo(HaveOccurred())

	By("Create cluster role")
	err = CreateClusterRole(gtc.KubeClient, "deploy/clusterRole.yaml")
	Expect(err).NotTo(HaveOccurred())

	By("Create cluster role binding")
	err = CreateClusterRoleBinding(gtc.KubeClient, gtc.Config.Namespace, "deploy/clusterRoleBinding.yaml")
	Expect(err).NotTo(HaveOccurred())

	By("Create deployment")
	err = CreateDeployment(
		gtc.KubeClient,
		gtc.Config.Namespace,
		gtc.Config.KubeDagImage,
		"deploy/deployment.yaml")

	Expect(err).NotTo(HaveOccurred())
}
