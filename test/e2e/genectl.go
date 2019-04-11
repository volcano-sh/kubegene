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
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	genev1alpha1 "kubegene.io/kubegene/pkg/apis/gene/v1alpha1"
	execclientset "kubegene.io/kubegene/pkg/client/clientset/versioned"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var GOPATH = os.Getenv("GOPATH")

var ToolRepo = filepath.Join(GOPATH, "src/kubegene.io/kubegene/example/tools")

var _ = DescribeGene("genectl", func(gtc *GeneTestContext) {
	var kubeClient kubernetes.Interface

	BeforeEach(func() {
		kubeClient = gtc.KubeClient
	})

	AfterEach(func() {

		By("Delete execution")

		list, err := gtc.GeneClient.ExecutionV1alpha1().Executions("default").List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())

		for i := 0; i < len(list.Items); i++ {
			err = gtc.GeneClient.ExecutionV1alpha1().Executions("default").Delete(list.Items[i].Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())

		}
	})

	It("sub single job", func() {
		By("Make dir for test")
		err := execCommand("mkdir", []string{"-p", "/tmp/subjob"})
		Expect(err).NotTo(HaveOccurred())

		By("Prepare shell script")
		err = execCommand("cp", []string{"example/single-job/print.sh", "/tmp/subjob/"})
		Expect(err).NotTo(HaveOccurred())

		createVolumeAndClaim("example/single-job/subjob-pv.yaml", "example/single-job/subjob-pvc.yaml", "default", kubeClient)

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

		createVolumeAndClaim("example/group-job/subrepjob-pv.yaml", "example/group-job/subrepjob-pvc.yaml", "default", kubeClient)

		By("Execute sub repjob command")
		cmd := NewGenectlCommand("sub", "repjob", "/tmp/subrepjob/repjob.sh", "--tool=nginx:latest", "--pvc=subrepjob-pvc", "--tool-repo="+ToolRepo)
		output := cmd.ExecOrDie()
		glog.Infof("output: %v", output)
	})

	It("sub workflow", func() {
		createVolumeAndClaim("example/simple-sample/sample-pv.yaml", "example/simple-sample/sample-pvc.yaml", "default", kubeClient)

		By("Execute sub workflow command")
		cmd := NewGenectlCommand("sub", "workflow", "example/simple-sample/simple-sample.yaml", "--tool-repo="+ToolRepo)
		output := cmd.ExecOrDie()
		glog.Infof("output: %v", output)
	})

	It("sub workflow with get_result", func() {
		createVolumeAndClaim("example/simple-sample-getresult/sample-pv.yaml", "example/simple-sample-getresult/sample-pvc.yaml", "default", kubeClient)

		By("Execute sub workflow command")
		cmd := NewGenectlCommand("sub", "workflow", "example/simple-sample-getresult/simple-sample-getresult.yaml", "--tool-repo="+ToolRepo)
		output := cmd.ExecOrDie()
		glog.Infof("output: %v", output)
		// wait to complete the execution
		glog.Infof("waiting to complete the execution")
		list, err := gtc.GeneClient.ExecutionV1alpha1().Executions("default").List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(1).To(Equal(len(list.Items)))

		err = WaitForExecutionSuccess(gtc.GeneClient, list.Items[0].Name, "default")
		Expect(err).NotTo(HaveOccurred())

		By("Check the result")

		result, err := ReadResultFrom("/kubegene-getresult/get-result.txt")

		Expect(err).NotTo(HaveOccurred())

		// The order of execution is variable, but it must be one of the following.
		expectResult := []string{
			"JOBMOD1JOBMOD2JOBMOD3JOBFINISH",
			"JOBMOD2JOBMOD3JOBMOD1JOBFINISH",
			"JOBMOD3JOBMOD1JOBMOD2JOBFINISH",
			"JOBMOD1JOBMOD3JOBMOD2JOBFINISH",
			"JOBMOD2JOBMOD1JOBMOD3JOBFINISH",
			"JOBMOD3JOBMOD2JOBMOD1JOBFINISH",
		}
		Expect(expectResult).Should(ContainElement(result))
	})

	It("sub workflow with check_result", func() {
		createVolumeAndClaim("example/conditional-sample/sample-pv.yaml", "example/conditional-sample/sample-pvc.yaml", "default", kubeClient)

		By("Execute sub workflow command")
		cmd := NewGenectlCommand("sub", "workflow", "example/conditional-sample/simple-sample-chkresult.yaml", "--tool-repo="+ToolRepo)
		output := cmd.ExecOrDie()
		glog.Infof("output: %v", output)
		// wait to complete the execution
		glog.Infof("waiting to complete the execution")
		list, err := gtc.GeneClient.ExecutionV1alpha1().Executions("default").List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(1).To(Equal(len(list.Items)))

		err = WaitForExecutionSuccess(gtc.GeneClient, list.Items[0].Name, "default")
		Expect(err).NotTo(HaveOccurred())

		By("Check the result")
		result, err := ReadResultFrom("/kubegene-chkresult/check-result.txt")

		Expect(err).NotTo(HaveOccurred())

		Expect(err).NotTo(HaveOccurred())
		// The order of execution is variable, but it must be one of the following.
		expectResult := []string{
			"CHK21CHK20JOBFINISH",
			"CHK20CHK21JOBFINISH",
		}
		Expect(expectResult).Should(ContainElement(result))
	})

	It("sub workflow with get_result and check_result", func() {
		createVolumeAndClaim("example/conditional-getresult-combination/sample-pv.yaml", "example/conditional-getresult-combination/sample-pvc.yaml", "default", kubeClient)

		By("Execute sub workflow command")
		cmd := NewGenectlCommand("sub", "workflow", "example/conditional-getresult-combination/simple-sample-get-chkresult.yaml", "--tool-repo="+ToolRepo)
		output := cmd.ExecOrDie()
		glog.Infof("output: %v", output)
		// wait to complete the execution
		glog.Infof("waiting to complete the execution")

		list, err := gtc.GeneClient.ExecutionV1alpha1().Executions("default").List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(1).To(Equal(len(list.Items)))

		err = WaitForExecutionSuccess(gtc.GeneClient, list.Items[0].Name, "default")
		Expect(err).NotTo(HaveOccurred())

		By("Check the result")
		result, err := ReadResultFrom("/kubegene-getchkresult/get-check-result.txt")

		Expect(err).NotTo(HaveOccurred())

		Expect(err).NotTo(HaveOccurred())
		// The order of execution is variable, but it must be one of the following.
		expectResult := []string{
			// 55 66  1  3
			"CHKRESULT55CHKRESULT66GETRESULT1GETRESULT3JOBBEFOREFINISHJOBFINISH",
			// 55 66  3  1
			"CHKRESULT55CHKRESULT66GETRESULT3GETRESULT1JOBBEFOREFINISHJOBFINISH",
			// 55  1  66   3
			"CHKRESULT55GETRESULT1CHKRESULT66GETRESULT3JOBBEFOREFINISHJOBFINISH",
			// 55  1  3   66
			"CHKRESULT55GETRESULT1GETRESULT3CHKRESULT66JOBBEFOREFINISHJOBFINISH",
			// 55  3  1   66
			"CHKRESULT55GETRESULT3GETRESULT1CHKRESULT66JOBBEFOREFINISHJOBFINISH",
			// 55  3  66   1
			"CHKRESULT55GETRESULT3CHKRESULT66GETRESULT1JOBBEFOREFINISHJOBFINISH",

			// 66  55   1    3
			"CHKRESULT66CHKRESULT55GETRESULT1GETRESULT3JOBBEFOREFINISHJOBFINISH",
			// 66  55   3    1
			"CHKRESULT66CHKRESULT55GETRESULT3GETRESULT1JOBBEFOREFINISHJOBFINISH",
			// 66  1    55   3
			"CHKRESULT66GETRESULT1CHKRESULT55GETRESULT3JOBBEFOREFINISHJOBFINISH",
			// 66  1    3    55
			"CHKRESULT66GETRESULT1GETRESULT3CHKRESULT55JOBBEFOREFINISHJOBFINISH",
			// 66  3    1    55
			"CHKRESULT66GETRESULT3GETRESULT1CHKRESULT55JOBBEFOREFINISHJOBFINISH",
			// 66  3    55   1
			"CHKRESULT66GETRESULT3CHKRESULT55GETRESULT1JOBBEFOREFINISHJOBFINISH",

			// 1 3  55  66
			"GETRESULT1GETRESULT3CHKRESULT55CHKRESULT66JOBBEFOREFINISHJOBFINISH",
			// 1 3  66  55
			"GETRESULT1GETRESULT3CHKRESULT66CHKRESULT55JOBBEFOREFINISHJOBFINISH",
			// 1 55  3  66
			"GETRESULT1CHKRESULT55GETRESULT3CHKRESULT66JOBBEFOREFINISHJOBFINISH",
			// 1 55  66  3
			"GETRESULT1CHKRESULT55CHKRESULT66GETRESULT3JOBBEFOREFINISHJOBFINISH",
			// 1 66  3   55
			"GETRESULT1CHKRESULT66GETRESULT3CHKRESULT55JOBBEFOREFINISHJOBFINISH",
			// 1 66  55  3
			"GETRESULT1CHKRESULT66CHKRESULT55GETRESULT3JOBBEFOREFINISHJOBFINISH",

			// 3  1   55  66
			"GETRESULT3GETRESULT1CHKRESULT55CHKRESULT66JOBBEFOREFINISHJOBFINISH",
			// 3  1   66  55
			"GETRESULT3GETRESULT1CHKRESULT66CHKRESULT55JOBBEFOREFINISHJOBFINISH",
			// 3  55  1   66
			"GETRESULT3CHKRESULT55GETRESULT1CHKRESULT66JOBBEFOREFINISHJOBFINISH",
			// 3  55  66   1
			"GETRESULT3CHKRESULT55CHKRESULT66GETRESULT1JOBBEFOREFINISHJOBFINISH",
			// 3  66  1   55
			"GETRESULT3CHKRESULT66GETRESULT1CHKRESULT55JOBBEFOREFINISHJOBFINISH",
			// 3  66  55   1
			"GETRESULT3CHKRESULT66CHKRESULT55GETRESULT1JOBBEFOREFINISHJOBFINISH",
		}
		Expect(expectResult).Should(ContainElement(result))
	})

	It("sub workflow with get_result and check_result in the same job", func() {
		createVolumeAndClaim("example/samejob-condition-getresult/sample-pv.yaml", "example/samejob-condition-getresult/sample-pvc.yaml", "default", kubeClient)

		By("Execute sub workflow command")
		cmd := NewGenectlCommand("sub", "workflow", "example/samejob-condition-getresult/samejob-condition-getresult.yaml", "--tool-repo="+ToolRepo)
		output := cmd.ExecOrDie()
		glog.Infof("output: %v", output)
		// wait to complete the execution
		glog.Infof("waiting to complete the execution")

		list, err := gtc.GeneClient.ExecutionV1alpha1().Executions("default").List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(1).To(Equal(len(list.Items)))

		err = WaitForExecutionSuccess(gtc.GeneClient, list.Items[0].Name, "default")
		Expect(err).NotTo(HaveOccurred())

		By("Check the result")
		result, err := ReadResultFrom("/kubegene-result/result.txt")

		Expect(err).NotTo(HaveOccurred())

		Expect(err).NotTo(HaveOccurred())
		// The order of execution is variable, but it must be one of the following.
		expectResult := []string{
			"ABC3ABC1JOBBEFOREFINISHJOBFINISH",
			"ABC1ABC3JOBBEFOREFINISHJOBFINISH",
		}
		Expect(expectResult).Should(ContainElement(result))
	})

	It("sub workflow with generic condition", func() {
		createVolumeAndClaim("example/generic-condition/sample-pv.yaml", "example/generic-condition/sample-pvc.yaml", "default", kubeClient)

		By("Execute sub workflow command")
		cmd := NewGenectlCommand("sub", "workflow", "example/generic-condition/generic-condition-workflow.yaml", "--tool-repo="+ToolRepo)
		output := cmd.ExecOrDie()
		glog.Infof("output: %v", output)
		// wait to complete the execution
		glog.Infof("waiting to complete the execution")

		list, err := gtc.GeneClient.ExecutionV1alpha1().Executions("default").List(metav1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(1).To(Equal(len(list.Items)))

		err = WaitForExecutionSuccess(gtc.GeneClient, list.Items[0].Name, "default")
		Expect(err).NotTo(HaveOccurred())

		By("Check the result")
		result, err := ReadResultFrom("/kubegene-generic/generic-condition.txt")

		Expect(err).NotTo(HaveOccurred())

		Expect(err).NotTo(HaveOccurred())
		// The order of execution is variable, but it must be one of the following.
		expectResult := []string{
			"ABC20ABC21JOBFINISH",
			"ABC21ABC20JOBFINISH",
		}
		Expect(expectResult).Should(ContainElement(result))
	})

	// test cases related to execution

	It("exec-1", func() {
		createVolumeAndClaim("example/execution/exec-pv.yaml", "example/execution/exec-pvc.yaml", "default", kubeClient)

		By("Execute exec-1")
		var execObj *genev1alpha1.Execution
		var err error
		execObj, err = createExecution("example/execution/exec-1.yaml", gtc.GeneClient)

		Expect(err).NotTo(HaveOccurred())
		Expect(execObj).NotTo(Equal(nil))

		err = WaitForExecutionSuccess(gtc.GeneClient, execObj.Name, "default")
		Expect(err).NotTo(HaveOccurred())

		By("Check the result")
		result, err := ReadResultFrom("/tmp/execution/exec-1.txt")

		Expect(err).NotTo(HaveOccurred())

		Expect(err).NotTo(HaveOccurred())
		// The order of execution is variable, but it must be one of the following.
		expectResult := []string{
			"ABCD",
			"ACBD",
		}
		Expect(expectResult).Should(ContainElement(result))
	})

	It("exec-2", func() {
		createVolumeAndClaim("example/execution/exec-pv.yaml", "example/execution/exec-pvc.yaml", "default", kubeClient)

		By("Execute exec-1")
		var execObj *genev1alpha1.Execution
		var err error
		execObj, err = createExecution("example/execution/exec-2.yaml", gtc.GeneClient)

		Expect(err).NotTo(HaveOccurred())
		Expect(execObj).NotTo(Equal(nil))

		err = WaitForExecutionSuccess(gtc.GeneClient, execObj.Name, "default")
		Expect(err).NotTo(HaveOccurred())

		By("Check the result")
		result, err := ReadResultFrom("/tmp/execution/exec-2.txt")

		Expect(err).NotTo(HaveOccurred())

		Expect(err).NotTo(HaveOccurred())
		// The order of execution is variable, but it must be one of the following.
		expectResult := []string{
			"AEBCDF",
			"AEBCFD",
			"AECBDF",
			"AECBFD",
			"EABCDF",
			"EABCFD",
			"EACBDF",
			"EACBFD",
			"AEBDCF",
		}
		Expect(expectResult).Should(ContainElement(result))
	})

	It("exec-5", func() {
		createVolumeAndClaim("example/execution/exec-pv.yaml", "example/execution/exec-pvc.yaml", "default", kubeClient)

		By("Execute exec-1")
		var execObj *genev1alpha1.Execution
		var err error
		execObj, err = createExecution("example/execution/exec-5.yaml", gtc.GeneClient)

		Expect(err).NotTo(HaveOccurred())
		Expect(execObj).NotTo(Equal(nil))

		err = WaitForExecutionSuccess(gtc.GeneClient, execObj.Name, "default")
		Expect(err).NotTo(HaveOccurred())

		By("Check the result")
		result, err := ReadResultFrom("/tmp/execution/exec-5-result.txt")

		Expect(err).NotTo(HaveOccurred())

		Expect(err).NotTo(HaveOccurred())
		// The order of execution is variable, but it must be one of the following.
		expectResult := []string{
			"CHKRESULTGETRESULT1GETRESULT3JOBBEFOREFINISHJOBFINISH",
			"CHKRESULTGETRESULT3GETRESULT1JOBBEFOREFINISHJOBFINISH",
			"GETRESULT1GETRESULT3CHKRESULTJOBBEFOREFINISHJOBFINISH",
			"GETRESULT3GETRESULT1CHKRESULTJOBBEFOREFINISHJOBFINISH",
			"GETRESULT1CHKRESULTGETRESULT3JOBBEFOREFINISHJOBFINISH",
			"GETRESULT3CHKRESULTGETRESULT1JOBBEFOREFINISHJOBFINISH",
		}
		Expect(expectResult).Should(ContainElement(result))
	})

	It("exec-6", func() {
		createVolumeAndClaim("example/execution/exec-pv.yaml", "example/execution/exec-pvc.yaml", "default", kubeClient)

		By("Execute exec-1")
		var execObj *genev1alpha1.Execution
		var err error
		execObj, err = createExecution("example/execution/exec-6.yaml", gtc.GeneClient)

		Expect(err).NotTo(HaveOccurred())
		Expect(execObj).NotTo(Equal(nil))

		err = WaitForExecutionSuccess(gtc.GeneClient, execObj.Name, "default")
		Expect(err).NotTo(HaveOccurred())

		By("Check the result")
		result, err := ReadResultFrom("/tmp/execution/exec-6-result.txt")

		Expect(err).NotTo(HaveOccurred())

		Expect(err).NotTo(HaveOccurred())
		// The order of execution is variable, but it must be one of the following.
		expectResult := []string{
			"GENERICCONDITIONJOBBEFOREFINISHJOBFINISH",
		}
		Expect(expectResult).Should(ContainElement(result))
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

func createExecution(execfile string, client execclientset.Interface) (*genev1alpha1.Execution, error) {
	By("Create a Execution")
	var execObj *genev1alpha1.Execution
	var exec *genev1alpha1.Execution
	var err error
	exec, err = executionFromYamlfile(execfile)
	Expect(err).NotTo(HaveOccurred())
	execObj, err = client.ExecutionV1alpha1().Executions("default").Create(exec)
	Expect(err).NotTo(HaveOccurred())
	return execObj, err

}

func execCommand(name string, args []string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	glog.Infof("command: %v, output: %v", name, string(output))
	return err
}

func ReadResultFrom(fullpath string) (string, error) {

	file, err := os.Open(fullpath)
	if err != nil {
		return "", fmt.Errorf("open file %s failed %v", fullpath, err)
	}
	defer file.Close()

	result := ""
	br := bufio.NewReader(file)
	for {
		bytes, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}
		output := string(bytes)
		result += output
	}
	return strings.TrimSpace(result), nil
}
