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
	"path/filepath"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	genev1alpha1 "kubegene.io/kubegene/pkg/apis/gene/v1alpha1"
	execclientset "kubegene.io/kubegene/pkg/client/clientset/versioned"
	"kubegene.io/kubegene/test/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const TestPath = "/kubegene"
const VolumeTimeout = 2 * time.Minute
const Poll = 2 * time.Second
const ExecutionTimeout = 5 * time.Minute

var _ = framework.DescribeGene("kube dag", func(gtc *framework.GeneTestContext) {
	var ns string
	var claimName string
	var pvName string
	var kubeClient kubernetes.Interface
	var geneClient execclientset.Interface

	BeforeEach(func() {
		ns = gtc.Config.Namespace
		kubeClient = gtc.KubeClient
		geneClient = gtc.GeneClient

		By("Create test volume")
		pv := makeTestPV("kubegene")
		pvName = pv.Name
		_, err := kubeClient.CoreV1().PersistentVolumes().Create(pv)
		Expect(err).NotTo(HaveOccurred())

		By("Create test claim")
		pvc := makeTestPVC(ns, "kubegene")
		claimName = pvc.Name
		_, err = kubeClient.CoreV1().PersistentVolumeClaims(ns).Create(pvc)
		Expect(err).NotTo(HaveOccurred())

		err = WaitForPersistentVolumeBound(kubeClient, pvName)
		Expect(err).NotTo(HaveOccurred())

		err = WaitForPersistentVolumeClaimBound(kubeClient, ns, claimName)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		By("Delete test claim")
		err := kubeClient.CoreV1().PersistentVolumeClaims(ns).Delete(claimName, &metav1.DeleteOptions{})
		Expect(err).NotTo(HaveOccurred())

		By("Delete test volume")
		err = kubeClient.CoreV1().PersistentVolumes().Delete(pvName, &metav1.DeleteOptions{})
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("run basic workflow", func() {
		It("whole depend type", func() {
			By("Create whole execution")
			execution := makeWholeExecution(ns, "kubegene", claimName)
			_, err := geneClient.ExecutionV1alpha1().Executions(ns).Create(execution)
			Expect(err).NotTo(HaveOccurred())

			err = WaitForExecutionSuccess(geneClient, execution.Name, ns)
			Expect(err).NotTo(HaveOccurred())

			By("Check execution result")
			result, err := ReadResult("whole.txt")
			Expect(err).NotTo(HaveOccurred())
			// The order of execution is variable, but it must be one of the following.
			expectResult := []string{"ABCD", "ACBD"}
			Expect(expectResult).Should(ConsistOf(result))

			By("Delete execution")
			err = geneClient.ExecutionV1alpha1().Executions(ns).Delete(execution.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("iterate depend type", func() {
			By("Create iterate execution")
			execution := makeIterateExecution(ns, "kubegene", claimName)
			_, err := geneClient.ExecutionV1alpha1().Executions(ns).Create(execution)
			Expect(err).NotTo(HaveOccurred())

			err = WaitForExecutionSuccess(geneClient, execution.Name, ns)
			Expect(err).NotTo(HaveOccurred())

			By("Check execution result")
			result, err := ReadResult("iterate.txt")
			Expect(err).NotTo(HaveOccurred())
			// The order of execution is variable, but it must be one of the following.
			expectResult := []string{
				"AB1B2C1C2",
				"AB1C1B2C2",
				"AB1B2C2C1",
				"AB2B1C2C1",
				"AB2C2B1C1",
				"AB2B1C1C2",
			}
			Expect(expectResult).Should(ConsistOf(result))

			By("Delete execution")
			err = geneClient.ExecutionV1alpha1().Executions(ns).Delete(execution.Name, &metav1.DeleteOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func makeWholeExecution(ns, prefix, pvc string) *genev1alpha1.Execution {
	exec := &genev1alpha1.Execution{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: prefix,
			Namespace:    ns,
		},
		Spec: genev1alpha1.ExecutionSpec{
			Tasks: []genev1alpha1.Task{
				{
					Name:       "A",
					Type:       genev1alpha1.JobTaskType,
					CommandSet: []string{"echo A >> /tmp/kubegene/whole.txt"},
					Image:      "busybox",
					Volumes: map[string]genev1alpha1.Volume{
						"volumeA": {
							MountPath: "/tmp/kubegene",
							MountFrom: genev1alpha1.VolumeSource{
								Pvc: pvc,
							},
						},
					},
				},
				{
					Name:       "B",
					Type:       genev1alpha1.JobTaskType,
					CommandSet: []string{"echo B >> /tmp/kubegene/whole.txt"},
					Image:      "busybox",
					Dependents: []genev1alpha1.Dependent{
						{
							Target: "A",
							Type:   genev1alpha1.DependTypeWhole,
						},
					},
					Volumes: map[string]genev1alpha1.Volume{
						"volumeB": {
							MountPath: "/tmp/kubegene",
							MountFrom: genev1alpha1.VolumeSource{
								Pvc: pvc,
							},
						},
					},
				},
				{
					Name:       "C",
					Type:       genev1alpha1.JobTaskType,
					CommandSet: []string{"echo C >> /tmp/kubegene/whole.txt"},
					Image:      "busybox",
					Dependents: []genev1alpha1.Dependent{
						{
							Target: "A",
							Type:   genev1alpha1.DependTypeWhole,
						},
					},
					Volumes: map[string]genev1alpha1.Volume{
						"volumeC": {
							MountPath: "/tmp/kubegene",
							MountFrom: genev1alpha1.VolumeSource{
								Pvc: pvc,
							},
						},
					},
				},
				{
					Name:       "D",
					Type:       genev1alpha1.JobTaskType,
					CommandSet: []string{"echo D >> /tmp/kubegene/whole.txt"},
					Image:      "busybox",
					Dependents: []genev1alpha1.Dependent{
						{
							Target: "B",
							Type:   genev1alpha1.DependTypeWhole,
						},
					},
					Volumes: map[string]genev1alpha1.Volume{
						"volumeD": {
							MountPath: "/tmp/kubegene",
							MountFrom: genev1alpha1.VolumeSource{
								Pvc: pvc,
							},
						},
					},
				},
			},
		},
	}

	return exec
}

func makeIterateExecution(ns, prefix, pvc string) *genev1alpha1.Execution {
	exec := &genev1alpha1.Execution{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: prefix,
			Namespace:    ns,
		},
		Spec: genev1alpha1.ExecutionSpec{
			Tasks: []genev1alpha1.Task{
				{
					Name:       "A",
					Type:       genev1alpha1.JobTaskType,
					CommandSet: []string{"echo A >> /tmp/kubegene/iterate.txt"},
					Image:      "busybox",
					Volumes: map[string]genev1alpha1.Volume{
						"volumeA": {
							MountPath: "/tmp/kubegene",
							MountFrom: genev1alpha1.VolumeSource{
								Pvc: pvc,
							},
						},
					},
				},
				{
					Name:       "B",
					Type:       genev1alpha1.JobTaskType,
					CommandSet: []string{"echo B1 >> /tmp/kubegene/iterate.txt", "echo B2 >> /tmp/kubegene/iterate.txt"},
					Image:      "busybox",
					Dependents: []genev1alpha1.Dependent{
						{
							Target: "A",
							Type:   genev1alpha1.DependTypeWhole,
						},
					},
					Volumes: map[string]genev1alpha1.Volume{
						"volumeB": {
							MountPath: "/tmp/kubegene",
							MountFrom: genev1alpha1.VolumeSource{
								Pvc: pvc,
							},
						},
					},
				},
				{
					Name:       "C",
					Type:       genev1alpha1.JobTaskType,
					CommandSet: []string{"echo C1 >> /tmp/kubegene/iterate.txt", "echo C2 >> /tmp/kubegene/iterate.txt"},
					Image:      "busybox",
					Dependents: []genev1alpha1.Dependent{
						{
							Target: "B",
							Type:   genev1alpha1.DependTypeIterate,
						},
					},
					Volumes: map[string]genev1alpha1.Volume{
						"volumeC": {
							MountPath: "/tmp/kubegene",
							MountFrom: genev1alpha1.VolumeSource{
								Pvc: pvc,
							},
						},
					},
				},
			},
		},
	}

	return exec
}

func makeTestPVC(ns, prefix string) *v1.PersistentVolumeClaim {
	claim := v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: prefix,
			Namespace:    ns,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse("1G"),
				},
			},
		},
	}
	return &claim
}

func makeTestPV(prefix string) *v1.PersistentVolume {
	hostPathType := v1.HostPathDirectoryOrCreate
	pv := v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: prefix,
		},
		Spec: v1.PersistentVolumeSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Capacity: v1.ResourceList{
				v1.ResourceStorage: resource.MustParse("1G"),
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: TestPath,
					Type: &hostPathType,
				},
			},
		},
	}

	return &pv
}

func WaitForPersistentVolumeBound(c kubernetes.Interface, pvName string) error {
	fmt.Printf("Waiting up to %v for PersistentVolume %s to have phase %s", VolumeTimeout, pvName, v1.VolumeBound)
	for start := time.Now(); time.Since(start) < VolumeTimeout; time.Sleep(Poll) {
		pv, err := c.CoreV1().PersistentVolumes().Get(pvName, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("Get persistent volume %s in failed, ignoring for %v: %v\n", pvName, Poll, err)
			continue
		} else {
			if pv.Status.Phase == v1.VolumeBound {
				fmt.Printf("PersistentVolume %s found and phase=%s (%v)\n", pvName, v1.VolumeBound, time.Since(start))
				return nil
			} else {
				fmt.Printf("PersistentVolume %s found but phase is %s instead of %s.\n", pvName, pv.Status.Phase, v1.VolumeBound)
			}
		}
	}
	return fmt.Errorf("PersistentVolume %s not in phase %s within %v\n", pvName, v1.VolumeBound, VolumeTimeout)
}

func WaitForPersistentVolumeClaimBound(c kubernetes.Interface, ns string, pvcName string) error {
	fmt.Printf("Waiting up to %v for PersistentVolumeClaim %s to have phase %s\n", VolumeTimeout, pvcName, v1.ClaimBound)
	for start := time.Now(); time.Since(start) < VolumeTimeout; time.Sleep(Poll) {
		pvc, err := c.CoreV1().PersistentVolumeClaims(ns).Get(pvcName, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("Failed to get claim %q, retrying in %v. Error: %v", pvcName, Poll, err)
			continue
		} else {
			if pvc.Status.Phase == v1.ClaimBound {
				fmt.Printf("PersistentVolumeClaim %s found and phase=%s (%v)\n", pvcName, v1.ClaimBound, time.Since(start))
				return nil
			} else {
				fmt.Printf("PersistentVolumeClaim %s found but phase is %s instead of %s.\n", pvcName, pvc.Status.Phase, v1.ClaimBound)
			}
		}
	}
	return fmt.Errorf("PersistentVolumeClaim %s not in phase %s within %v\n", pvcName, v1.ClaimBound, VolumeTimeout)
}

func WaitForExecutionSuccess(client execclientset.Interface, name, ns string) error {
	fmt.Printf("Waiting up to %v for Execution %s to be successed\n", ExecutionTimeout, name)
	for start := time.Now(); time.Since(start) < ExecutionTimeout; time.Sleep(Poll) {
		execution, err := client.ExecutionV1alpha1().Executions(ns).Get(name, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("Failed to get execution %q, retrying in %v. Error: %v", name, Poll, err)
			continue
		} else {
			if execution.Status.Phase == genev1alpha1.VertexSucceeded {
				fmt.Printf("Execution %s found and phase=%s (%v)\n", name, genev1alpha1.VertexSucceeded, time.Since(start))
				return nil
			} else {
				fmt.Printf("Execution %s found but phase is %s instead of %s.\n", name, execution.Status.Phase, genev1alpha1.VertexSucceeded)
			}
		}
	}
	return fmt.Errorf("Execution %s not in phase %s within %v\n", name, genev1alpha1.VertexSucceeded, ExecutionTimeout)
}

func ReadResult(fileName string) (string, error) {
	path := filepath.Join(TestPath, fileName)
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open file %s failed %v", path, err)
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
