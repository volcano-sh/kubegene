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

package app

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	apiv1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"

	"kubegene.io/kubegene/cmd/kube-dag/app/options"
	"kubegene.io/kubegene/pkg/apis/gene"
	genev1alpha1 "kubegene.io/kubegene/pkg/apis/gene/v1alpha1"
	execclientset "kubegene.io/kubegene/pkg/client/clientset/versioned"
	execscheme "kubegene.io/kubegene/pkg/client/clientset/versioned/scheme"
	execinformers "kubegene.io/kubegene/pkg/client/informers/externalversions"
	"kubegene.io/kubegene/pkg/controller"
	"kubegene.io/kubegene/pkg/util"
	"kubegene.io/kubegene/pkg/version"
)

const (
	leaseDuration = 15 * time.Second
	renewDeadline = 10 * time.Second
	retryPeriod   = 2 * time.Second
)

func createRecorder(kubeClient clientset.Interface) record.EventRecorder {
	// Add Execution types to the defualt Kubernetes so events can be logged for them.
	execscheme.AddToScheme(scheme.Scheme)
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: v1core.New(kubeClient.CoreV1().RESTClient()).Events("")})
	return eventBroadcaster.NewRecorder(scheme.Scheme, apiv1.EventSource{Component: "gene-controller"})
}

func createClients(o *options.ExecutionOption) (clientset.Interface, clientset.Interface, *execclientset.Clientset, *apiextensionsclient.Clientset, error) {
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", o.KubeConfig)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("unable to build config from flags: %v", err)
	}

	// Override kubeconfig qps/burst settings from flags
	kubeconfig.QPS = o.KubeAPIQPS
	kubeconfig.Burst = int(o.KubeAPIBurst)

	kubeClient, err := clientset.NewForConfig(restclient.AddUserAgent(kubeconfig, "gene-controller"))
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	leaderElectionClient := clientset.NewForConfigOrDie(restclient.AddUserAgent(kubeconfig, "leader-election"))

	apiextentionsClient, err := apiextensionsclient.NewForConfig(kubeconfig)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	geneClient, err := execclientset.NewForConfig(kubeconfig)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	return kubeClient, leaderElectionClient, geneClient, apiextentionsClient, nil
}

func installExecutionCRD(apiextensionsclient apiextensionsclient.Interface) error {
	crd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: gene.ExecutionPlural + "." + gene.GroupName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   gene.GroupName,
			Version: genev1alpha1.SchemeGroupVersion.Version,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural: gene.ExecutionPlural,
				Kind:   reflect.TypeOf(genev1alpha1.Execution{}).Name(),
			},
			Subresources: &apiextensionsv1beta1.CustomResourceSubresources{
				Status: &apiextensionsv1beta1.CustomResourceSubresourceStatus{},
			},
		},
	}

	if err := util.EnsureCreateCRD(apiextensionsclient, crd); err != nil {
		return err
	}

	return nil
}

func Run(o *options.ExecutionOption, stopCh <-chan struct{}) error {
	if o.PrintVersion {
		version := version.GetVersion()
		fmt.Printf("  kube-dag Version: %s\n", version)
		os.Exit(0)
	}
	kubeClient, leaderElectionClient, geneClient, apiextentionsClient, err := createClients(o)
	if err != nil {
		return err
	}

	// ensure execution resource has been created, if not, create it.
	if err := installExecutionCRD(apiextentionsClient); err != nil {
		return err
	}

	sharedInformers := informers.NewSharedInformerFactory(kubeClient, o.ResyncPeriod)
	geneInformer := execinformers.NewSharedInformerFactory(geneClient, o.ResyncPeriod)
	eventRecorder := createRecorder(kubeClient)
	parameter := &controller.ControllerParameters{
		EventRecorder:     eventRecorder,
		KubeClient:        kubeClient,
		ExecutionClient:   geneClient.ExecutionV1alpha1(),
		JobInformer:       sharedInformers.Batch().V1().Jobs(),
		ExecutionInformer: geneInformer.Execution().V1alpha1().Executions(),
	}

	execCtrl := controller.NewExecutionController(parameter)
	run := func(ctx context.Context) {
		go sharedInformers.Start(stopCh)
		go geneInformer.Start(stopCh)
		execCtrl.Run(1, stopCh)
		<-stopCh
	}

	if !o.LeaderElect {
		run(context.TODO())
		panic("unreachable")
	}

	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("unable to get hostname: %v", err)
	}
	// add a uniquifier so that two processes on the same host don't accidentally both become active
	id := hostname + "_" + string(uuid.NewUUID())

	rl, err := resourcelock.New(resourcelock.ConfigMapsResourceLock,
		o.LockObjectNamespace,
		"kubegene-controller",
		leaderElectionClient.CoreV1(),
		leaderElectionClient.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: eventRecorder,
		})
	if err != nil {
		return fmt.Errorf("couldn't create resource lock: %v", err)
	}

	leaderelection.RunOrDie(context.TODO(), leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: leaseDuration,
		RenewDeadline: renewDeadline,
		RetryPeriod:   retryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: run,
			OnStoppedLeading: func() {
				klog.Fatalf("leaderelection lost")
			},
		},
	})
	panic("unreachable")
}
