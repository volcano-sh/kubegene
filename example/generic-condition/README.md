## Overview

This is a simple example to demonstrate how to submit a workflow using `genectl` with generic condition 

## Prerequisites

 * Create the volume and claim.
   ```
   $ kubectl create -f sample-pv.yaml
   $ kubectl create -f sample-pvc.yaml
   ```
 * Ensure your tool repo has been set correctly.

## Command

```bash
$ genectl sub workflow generic-condition-workflow.yaml
```


