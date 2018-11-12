## Overview

This is a simple example to demonstrate how to submit a group job using `genectl`

## Prerequisites

 * Create directory `/tmp/subrepjob`
 * Move the shell script `print-date.sh`, `print-number.sh` and `repjob.sh` to directory `/tmp/subrepjob`.
 * Create the volume and claim.
   ```
   $ kubectl create -f subrepjob-pv.yaml
   $ kubectl create -f subrepjob-pvc.yaml
   ```
 * Ensure your tool repo has been set correctly.

## Command

```bash
$ gcs sub job /tmp/subrepjob/repjob.sh  --tool nginx:latest --pvc subrepjob-pvc
```