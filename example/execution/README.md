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

The below example is with the nfs

## Prerequisites

 * k8s is deployed with minikube in a machine /laptop.
 * kubegene is deployed
 * deploy the nfs-provisioner on the same machine/laptop as explained below

     just clone and build the project
     $ cd $GOPATH/src/github.com/kubernetes-incubator
     $ git clone https://github.com/kubernetes-incubator/external-storage.git
     $ cd $GOPATH/src/github.com/kubernetes-incubator/external-storage/nfs
     $ make
     brought up the nfs-provisioner as shown below (supposed to create /export folder in that machine where all the data is stored)
     $ sudo ./nfs-provisioner -provisioner=example.com/nfs
     -kubeconfig=$HOME/.kube/config
     -run-server=false
     -use-ganesha=false


   ```bash
   $ kubectl create -f nfs-storageclass.yaml
   Note: we can verify that the StorageClass is created or not by  kubectl get StorageClass
   $ kubectl create -f nfs-pvc.yaml
   Note: we can verify that the pv & pvc are created or not by  kubectl get pvc and kubectl get pv
   ```

## Command

```bash
$ kubectl create -f exec-1-nfs.yaml

```
