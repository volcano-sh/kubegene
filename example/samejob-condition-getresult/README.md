## Overview

This is a simple example to demonstrate how to submit a workflow using `genectl` same job has condition/check_result and get_result  

## Prerequisites

 * Create the volume and claim.
   ```
   $ kubectl create -f sample-pv.yaml
   $ kubectl create -f sample-pvc.yaml
   ```
 * Ensure your tool repo has been set correctly.

## Command

```bash
$ genectl sub workflow samejob-condition-getresult.yaml
```

[MoreInfo](https://kubegene.io/docs/design/conditional-concurrency/conditional-concurrency.md)

[MoreInfo](https://kubegene.io/docs/design/dynamic-concurrency/dynamic-concurrency.md)
