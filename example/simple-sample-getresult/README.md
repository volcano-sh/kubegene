## Overview

This is a simple example to demonstrate how to submit a workflow using `genectl` with get_result

## Prerequisites

 * Create the volume and claim.
   ```
   $ kubectl create -f sample-pv.yaml
   $ kubectl create -f sample-pvc.yaml
   ```
 * Ensure your tool repo has been set correctly.

## Command

```bash
$ genectl sub workflow simple-sample-getresult.yaml
```


MoreInfo](https://kubegene.io/docs/design/dynamic-concurrency/dynamic-concurrency.md)
