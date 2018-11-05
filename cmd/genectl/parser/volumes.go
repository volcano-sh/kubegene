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

package parser

import (
	"fmt"
	execv1alpha1 "kubegene.io/kubegene/pkg/apis/gene/v1alpha1"
	"path"
)

func ValidateVolumes(volumes map[string]Volume, inputs map[string]Input) ErrorList {
	errors := ErrorList{}
	for key, volume := range volumes {
		pvc := volume.MountFrom.PVC
		if len(pvc) == 0 {
			err := fmt.Errorf("volumes[%s].mountFrom: volume only support pvc and mountFrom.pvc should not be empty", key)
			errors = append(errors, err)
			continue
		}
		if IsVariant(pvc) {
			prefix := fmt.Sprintf("volumes[%s].mountFrom.pvc", key)
			if err := ValidateVariant(prefix, pvc, []string{StringType}, inputs); err != nil {
				errors = append(errors, err)
			}
		}

		mountPath := volume.MountPath
		if len(mountPath) == 0 {
			errors = append(errors, fmt.Errorf("volumes[%s].mountPath: mountPath should be empty", key))
			continue
		} else if IsVariant(mountPath) {
			prefix := fmt.Sprintf("volumes[%s].mountPath", key)
			if err := ValidateVariant(prefix, mountPath, []string{StringType}, inputs); err != nil {
				errors = append(errors, err)
			}
		} else if !path.IsAbs(mountPath) {
			err := fmt.Errorf("volumes[%s].mountPath: mountPath should be an absolute path, but the real one is %s", key, mountPath)
			errors = append(errors, err)
		}
	}
	return errors
}

func TransVolume2ExecVolume(volumes map[string]Volume) map[string]execv1alpha1.Volume {
	execVolumes := make(map[string]execv1alpha1.Volume, len(volumes))
	for key, volume := range volumes {
		var tmpVolume execv1alpha1.Volume
		tmpVolume.MountFrom.Pvc = volume.MountFrom.PVC
		tmpVolume.MountPath = volume.MountPath
		execVolumes[key] = tmpVolume
	}
	return execVolumes
}
