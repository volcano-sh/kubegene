## Overview

This is a simple example to demonstrate how to submit a single job using `genectl`

## Prerequisites

 * Create directory `/tmp/subjob`
 * Move the shell script `print.sh` to directory `/tmp/subjob`.
 * Create the volume and claim.
   ```
   $ kubectl create -f subjob-pv.yaml
   $ kubectl create -f subjob-pvc.yaml
   ```
 * Ensure your tool repo has been set correctly.

## Command

```bash
$ genectl sub job /tmp/subjob/print.sh  --tool nginx:latest --pvc subjob-pvc
```