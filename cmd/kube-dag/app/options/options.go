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

package options

import (
	"time"

	"github.com/spf13/pflag"
)

const (
	defaultKubeAPIQPS   = 20.0
	defaultKubeAPIBurst = 30
)

type ExecutionOption struct {
	KubeConfig   string
	KubeAPIQPS   float32
	KubeAPIBurst int32
	LeaderElect  bool
	// the namespace of the lock object
	LockObjectNamespace string
	ResyncPeriod        time.Duration
	PrintVersion        bool
}

func NewExecutionOption() *ExecutionOption {
	// init default configuration
	return &ExecutionOption{
		KubeAPIQPS:          defaultKubeAPIQPS,
		KubeAPIBurst:        defaultKubeAPIBurst,
		LeaderElect:         true,
		LockObjectNamespace: "kube-system",
		ResyncPeriod:        60 * time.Second,
		PrintVersion:        false,
	}
}

func (o *ExecutionOption) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}

	fs.StringVar(&o.KubeConfig, "kubeconfig", o.KubeConfig, "Path to kubeconfig file with authorization and master location information.")
	fs.Float32Var(&o.KubeAPIQPS, "kube-api-qps", o.KubeAPIQPS, "QPS to use while talking with kubernetes apiserver.")
	fs.Int32Var(&o.KubeAPIBurst, "kube-api-burst", o.KubeAPIBurst, "Burst to use while talking with kubernetes apiserver.")
	fs.DurationVar(&o.ResyncPeriod, "resyncPeriod", o.ResyncPeriod, "The period that should be used to re-sync the execution.")
	fs.BoolVar(&o.LeaderElect, "leader-elect", o.LeaderElect, "Start a leader election client and gain leadership before executing the main loop.")
	fs.StringVar(&o.LockObjectNamespace, "lock-object-namespace", o.LockObjectNamespace, "The namespace of the lock object.")
	fs.BoolVar(&o.PrintVersion, "version", o.PrintVersion, "Show version and quit")
}
