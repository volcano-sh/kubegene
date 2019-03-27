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
	"os"
	"os/exec"
	"path/filepath"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var GOPATH = os.Getenv("GOPATH")

var ToolRepo = filepath.Join(GOPATH, "src/kubegene.io/kubegene/example/tools")

var _ = DescribeGene("genectl", func(gtc *GeneTestContext) {
	var ns string
	var kubeClient kubernetes.Interface

	BeforeEach(func() {
		ns = gtc.Config.Namespace
		kubeClient = gtc.KubeClient
	})

	AfterEach(func() {
	})

	It("sub single job", func() {
		By("Make dir for test")
		err := execCommand("mkdir", []string{"-p", "/tmp/subjob"})
		Expect(err).NotTo(HaveOccurred())

		By("Prepare shell script")
		err = execCommand("cp", []string{"example/single-job/print.sh", "/tmp/subjob/"})
		Expect(err).NotTo(HaveOccurred())

		createVolumeAndClaim("example/single-job/subjob-pv.yaml", "example/single-job/subjob-pvc.yaml", ns, kubeClient)

		By("Execute sub job command")
		cmd := NewGenectlCommand("sub", "job", "/tmp/subjob/print.sh", "--tool=nginx:latest", "--pvc=subjob-pvc", "--tool-repo="+ToolRepo)
		output := cmd.ExecOrDie()
		glog.Infof("output: %v", output)
	})

	It("sub group job", func() {
		By("Make dir for test")
		err := execCommand("mkdir", []string{"-p", "/tmp/subrepjob"})
		Expect(err).NotTo(HaveOccurred())

		By("Prepare shell script")
		err = execCommand("cp", []string{"example/group-job/print-date.sh", "/tmp/subrepjob/"})
		Expect(err).NotTo(HaveOccurred())
		err = execCommand("cp", []string{"example/group-job/print-number.sh", "/tmp/subrepjob/"})
		Expect(err).NotTo(HaveOccurred())
		err = execCommand("cp", []string{"example/group-job/repjob.sh", "/tmp/subrepjob/"})
		Expect(err).NotTo(HaveOccurred())

		createVolumeAndClaim("example/group-job/subrepjob-pv.yaml", "example/group-job/subrepjob-pvc.yaml", ns, kubeClient)

		By("Execute sub repjob command")
		cmd := NewGenectlCommand("sub", "repjob", "/tmp/subrepjob/repjob.sh", "--tool=nginx:latest", "--pvc=subrepjob-pvc", "--tool-repo="+ToolRepo)
		output := cmd.ExecOrDie()
		glog.Infof("output: %v", output)
	})

	It("sub workflow", func() {
		createVolumeAndClaim("example/simple-sample/sample-pv.yaml", "example/simple-sample/sample-pvc.yaml", ns, kubeClient)

		By("Execute sub workflow command")
		cmd := NewGenectlCommand("sub", "workflow", "example/simple-sample/simple-sample.yaml", "--tool-repo="+ToolRepo)
		output := cmd.ExecOrDie()
		glog.Infof("output: %v", output)
	})

	It("sub workflow with get_result", func() {
		createVolumeAndClaim("example/simple-sample-getresult/sample-pv.yaml", "example/simple-sample-getresult/sample-pvc.yaml", ns, kubeClient)

		By("Execute sub workflow command")
		cmd := NewGenectlCommand("sub", "workflow", "example/simple-sample-getresult/simple-sample-getresult.yaml", "--tool-repo="+ToolRepo)
		output := cmd.ExecOrDie()
		glog.Infof("output: %v", output)
	})
})

func createVolumeAndClaim(volumeFile, claimFile, ns string, kubeClient kubernetes.Interface) {
	By("Create a volume")
	volume, err := volumeFromManifest(volumeFile)
	Expect(err).NotTo(HaveOccurred())
	_, err = kubeClient.CoreV1().PersistentVolumes().Create(volume)
	Expect(err == nil || errors.IsAlreadyExists(err)).To(Equal(true))

	By("Create a claim")
	claim, err := claimFromManifest(claimFile)
	Expect(err).NotTo(HaveOccurred())
	claim.Namespace = ns
	_, err = kubeClient.CoreV1().PersistentVolumeClaims(ns).Create(claim)
	Expect(err == nil || errors.IsAlreadyExists(err)).To(Equal(true))

	err = WaitForPersistentVolumeBound(kubeClient, volume.Name)
	Expect(err).NotTo(HaveOccurred())

	err = WaitForPersistentVolumeClaimBound(kubeClient, ns, claim.Name)
	Expect(err).NotTo(HaveOccurred())
}

func execCommand(name string, args []string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	glog.Infof("command: %v, output: %v", name, string(output))
	return err
}
