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
	"fmt"
	"time"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"io/ioutil"
	apps "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"path/filepath"
)

func CreateServiceAccount(kubeClient kubernetes.Interface, namespace string, relativePath string) error {
	serviceAccount, err := serviceAccountFromManifest(relativePath, namespace)
	if err != nil {
		return err
	}

	_, err = kubeClient.CoreV1().ServiceAccounts(namespace).Create(serviceAccount)
	return err
}

func DeleteServiceAccount(kubeClient kubernetes.Interface, namespace string, relativePath string) error {
	serviceAccount, err := serviceAccountFromManifest(relativePath, namespace)
	if err != nil {
		return err
	}

	return kubeClient.CoreV1().ServiceAccounts(namespace).Delete(serviceAccount.Name, nil)
}

func serviceAccountFromManifest(fileName, ns string) (*v1.ServiceAccount, error) {
	data, err := readFile(fileName)
	if err != nil {
		return nil, err
	}

	var serviceAccount v1.ServiceAccount
	err = yaml.Unmarshal(data, &serviceAccount)
	if err != nil {
		return nil, err
	}

	serviceAccount.Namespace = ns

	return &serviceAccount, nil
}

func CreateClusterRole(kubeClient kubernetes.Interface, relativePath string) error {
	clusterRole, err := clusterRoleFromManifest(relativePath)
	if err != nil {
		return err
	}

	kubeClient.RbacV1().ClusterRoles().Delete(clusterRole.GetName(), &metav1.DeleteOptions{})

	err = wait.Poll(2*time.Second, 10*time.Minute, func() (bool, error) {
		_, err := kubeClient.RbacV1().ClusterRoles().Get(clusterRole.GetName(), metav1.GetOptions{})
		return apierrs.IsNotFound(err), nil
	})
	if err != nil {
		return err
	}

	_, err = kubeClient.RbacV1().ClusterRoles().Create(clusterRole)
	return err
}

func DeleteClusterRole(kubeClient kubernetes.Interface, relativePath string) error {
	clusterRole, err := clusterRoleFromManifest(relativePath)
	if err != nil {
		return err
	}

	return kubeClient.RbacV1().ClusterRoles().Delete(clusterRole.Name, nil)
}

func clusterRoleFromManifest(fileName string) (*rbac.ClusterRole, error) {
	data, err := readFile(fileName)
	if err != nil {
		return nil, err
	}

	var clusterRole rbac.ClusterRole
	err = yaml.Unmarshal(data, &clusterRole)
	if err != nil {
		return nil, err
	}

	return &clusterRole, nil
}

func CreateClusterRoleBinding(kubeClient kubernetes.Interface, ns, relativePath string) error {
	clusterRoleBinding, err := clusterRoleBindingFromManifest(relativePath)
	if err != nil {
		return err
	}

	if len(ns) != 0 {
		clusterRoleBinding.Subjects[0].Namespace = ns
	}

	kubeClient.RbacV1().ClusterRoleBindings().Delete(clusterRoleBinding.GetName(), &metav1.DeleteOptions{})

	err = wait.Poll(2*time.Second, 10*time.Minute, func() (bool, error) {
		_, err := kubeClient.RbacV1().ClusterRoleBindings().Get(clusterRoleBinding.GetName(), metav1.GetOptions{})
		return apierrs.IsNotFound(err), nil
	})
	if err != nil {
		return err
	}

	_, err = kubeClient.RbacV1().ClusterRoleBindings().Create(clusterRoleBinding)
	return err
}

func DeleteClusterRoleBinding(kubeClient kubernetes.Interface, relativePath string) error {
	clusterRoleBinding, err := clusterRoleBindingFromManifest(relativePath)
	if err != nil {
		return err
	}

	return kubeClient.RbacV1().ClusterRoleBindings().Delete(clusterRoleBinding.Name, nil)
}

func clusterRoleBindingFromManifest(fileName string) (*rbac.ClusterRoleBinding, error) {
	data, err := readFile(fileName)
	if err != nil {
		return nil, err
	}

	var clusterRoleBinding rbac.ClusterRoleBinding
	err = yaml.Unmarshal(data, &clusterRoleBinding)
	if err != nil {
		return nil, err
	}

	return &clusterRoleBinding, nil
}

func CreateDeployment(kubeClient kubernetes.Interface, ns, image, relativePath string) error {
	deployment, err := deploymentFromManifest(relativePath)
	if err != nil {
		return err
	}

	deployment.Namespace = ns
	if len(image) != 0 {
		deployment.Spec.Template.Spec.Containers[0].Image = image
	}

	newDeployment, err := kubeClient.AppsV1().Deployments(ns).Create(deployment)
	if err != nil {
		return err
	}

	return waitDeploymentReady(kubeClient, newDeployment)
}

func DeleteDeployment(kubeClient kubernetes.Interface, ns, relativePath string) error {
	deployment, err := deploymentFromManifest(relativePath)
	if err != nil {
		return err
	}

	err = kubeClient.AppsV1().Deployments(ns).Delete(deployment.Name, nil)
	if err != nil {
		return err
	}

	selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return err
	}
	options := metav1.ListOptions{LabelSelector: selector.String()}

	// Ensuring deployment Pods were deleted
	var pods *v1.PodList
	if err := wait.PollImmediate(time.Second, 3*time.Minute, func() (bool, error) {
		pods, err = kubeClient.CoreV1().Pods(ns).List(options)
		if err != nil {
			return false, err
		}

		if len(pods.Items) == 0 {
			return true, nil
		}
		return false, nil
	}); err != nil {
		return fmt.Errorf("Err : %s\n. Failed to remove deployment %s pods : %+v", err, deployment.Name, pods)
	}

	return nil
}

func deploymentFromManifest(fileName string) (*apps.Deployment, error) {
	data, err := readFile(fileName)
	if err != nil {
		return nil, err
	}

	var deployment apps.Deployment
	err = yaml.Unmarshal(data, &deployment)
	if err != nil {
		return nil, err
	}

	return &deployment, nil
}

func waitDeploymentReady(kubeClient kubernetes.Interface, d *apps.Deployment) error {
	var deployment *apps.Deployment
	err := wait.PollImmediate(3*time.Second, 3*time.Minute, func() (bool, error) {
		var err error
		deployment, err = kubeClient.AppsV1().Deployments(d.Namespace).Get(d.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		// When the deployment status and its underlying resources reach the desired state, we're done
		if DeploymentComplete(d, &deployment.Status) {
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return fmt.Errorf("error waiting for deployment %q status to match expectation: %v", d.Name, err)
	}

	return nil
}

func DeploymentComplete(deployment *apps.Deployment, newStatus *apps.DeploymentStatus) bool {
	glog.Infof("Number of ready pod %v", newStatus.ReadyReplicas)
	return newStatus.ReadyReplicas == *(deployment.Spec.Replicas)
}

func readFile(path string) ([]byte, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(absPath)
	return data, err
}

func volumeFromManifest(fileName string) (*v1.PersistentVolume, error) {
	data, err := readFile(fileName)
	if err != nil {
		return nil, err
	}

	var persistentVolume v1.PersistentVolume
	err = yaml.Unmarshal(data, &persistentVolume)
	if err != nil {
		return nil, err
	}

	return &persistentVolume, nil
}

func claimFromManifest(fileName string) (*v1.PersistentVolumeClaim, error) {
	data, err := readFile(fileName)
	if err != nil {
		return nil, err
	}

	var claim v1.PersistentVolumeClaim
	err = yaml.Unmarshal(data, &claim)
	if err != nil {
		return nil, err
	}

	return &claim, nil
}
