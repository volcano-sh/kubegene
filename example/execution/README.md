## Overview

There are some examples to demonstrate the kubedag engine how to execute various execution.

## Prerequisites

 * Create the volume and claim.
   ```
   $ kubectl create -f exec-pv.yaml
   $ kubectl create -f exec-pvc.yaml
   ```

## Command

```bash
$ kubectl create -f exec-1.yaml
$ kubectl create -f exec-2.yaml
$ kubectl create -f exec-3.yaml
$ kubectl create -f iterate-exec.yaml
```
