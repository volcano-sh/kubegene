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
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/golang/glog"
	. "github.com/onsi/gomega"
)

type genectlWrapper struct {
	cmd     *exec.Cmd
	timeout <-chan time.Time
}

func NewGenectlCommand(args ...string) *genectlWrapper {
	ctlWrapper := new(genectlWrapper)
	ctlWrapper.cmd = GenectlCmd(args...)
	return ctlWrapper
}

// GenectlCmd runs the genectl executable through the wrapper script.
func GenectlCmd(args ...string) *exec.Cmd {
	defaultArgs := []string{}

	if geneTestContext.Config.KubeConfig != "" {
		defaultArgs = append(defaultArgs, "--kubeconfig="+geneTestContext.Config.KubeConfig)
	}
	genectlArgs := append(defaultArgs, args...)

	//We allow users to specify path to genectl.
	cmd := exec.Command(geneTestContext.Config.GenectlPath, genectlArgs...)

	//caller will invoke this and wait on it.
	return cmd
}

func (ctlWrapper genectlWrapper) ExecOrDie() string {
	str, err := ctlWrapper.Exec()
	Expect(err).NotTo(HaveOccurred())
	return str
}

func (ctlWrapper genectlWrapper) Exec() (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := ctlWrapper.cmd
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	glog.Infof("Running '%s %s'", cmd.Path, strings.Join(cmd.Args, " "))
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("error starting %v:\nCommand stdout:\n%v\nstderr:\n%v\nerror:\n%v\n", cmd, cmd.Stdout, cmd.Stderr, err)
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Wait()
	}()
	select {
	case err := <-errCh:
		if err != nil {
			var rc int = 127
			if ee, ok := err.(*exec.ExitError); ok {
				rc = int(ee.Sys().(syscall.WaitStatus).ExitStatus())
				glog.Infof("rc: %d", rc)
			}
			return "", fmt.Errorf("error running %v:\nCommand stdout:\n%v\nstderr:\n%v\nerror:\n%v\n", cmd, cmd.Stdout, cmd.Stderr, err)
		}
	case <-ctlWrapper.timeout:
		ctlWrapper.cmd.Process.Kill()
		return "", fmt.Errorf("timed out waiting for command %v:\nCommand stdout:\n%v\nstderr:\n%v\n", cmd, cmd.Stdout, cmd.Stderr)
	}
	glog.Infof("stderr: %q", stderr.String())
	glog.Infof("stdout: %q", stdout.String())
	return stdout.String(), nil
}
