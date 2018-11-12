# KubeGene
[![Go Report Card](https://goreportcard.com/badge/github.com/kubegene/kubegene)](https://goreportcard.com/report/github.com/kubegene/kubegene)

<img src="./images/kubegene_logo.png">

## What is KubeGene
The KubeGene is dedicated to making genome sequencing process simple, portable and scalable. It provides a complete solution for genome sequencing on the kubernetes and supports mainstream biological genome sequencing scenarios such as DNA, RNA, and liquid biopsy. KubeGene is based on lightweight container technology and official standard algorithms. You can make a flexible and customizable genome sequencing process by using KubeGene.

## Advantages

### Universal workflow design grammar
KubeGene provides a complete set of genome sequencing workflow grammars which decouples with specific analysis tools. It requires a very low learning cost to learn how to write and use the workflow. You can easily migrate the genome sequencing business to KubeGene.

### Tailor-made workflow for the biosequencing industry
The workflow grammar is designed by comparing different genome sequencing scenarios. It also keeps the user’s traditional usage habit as much as possible and is closer to user scenarios.

### More efficient resource usage
KubeGene uses container to run the genome sequencing business. Compared to traditional genome sequencing solutions using virtual machines, KubeGene makes resource usage more efficient and avoiding resource idle.

### Scaling based on demand
Kubernetes can automatically scale your cluster based on your genome sequencing workload by using KubeGene. Also you can easily scale the kubernetes cluster manually. It can save your production costs. 


## Prerequisites


- Kubernetes 1.10+
- golang 1.10+

If you have an older version of Kubernetes, we recommend upgrading Kubernetes first and then use the kubeGene.  
If you use 1.10 kubernetes, you should make sure that the feature gate `CustomResourceSubresources` is open.  
You can use minikube to build a local kubernetes cluster quickly. For more information, you can use [minikube](https://kubegene.netlify.com/docs/started/getting-started-minikube/).

## Deploy KubeGene

KubeGene has two main componments.

* **Kubedag**. Kubedag is a DAG workflow engine on Kubernetes platform. It is dedicated to making Gene Sequencing workflow execute in container easily. 

* **genectl**. genectl is a command line interface for running commands against KubeGene. 

### Clone code
First, make work directory:
```bash
$ mkdir -p $GOPATH/src/kubegene.io
```
Enter the work directory and clone the code:
```bash
$ git clone https://github.com/kubegene/kubegene.git
```
### Install kubedag

We have provide the YAML file that contains all API objects that are necessary to run kubedag, you can easily deploy the kubedag using the follow commands:

```bash
$ kubectl create -f deploy/setup-kubedag.yaml
```

kubedag will automatically create a Kubernetes Custom Resource Definition (CRD):

```bash
$ kubectl get customresourcedefinitions
NAME                                    AGE
executions.execution.kubegene.io        2s 
```
For more information, you can see [Kugedag](https://kubegene.netlify.com/docs/about/kubedag/).

### Install genectl

Compile the `genectl`

```bash
$ make genectl
```
Enter bin directory and make the genectl binary executable.
```bash
$ chmod +x ./genectl
```
Move the binary in to your PATH.
```bash
$ sudo mv ./genectl /usr/local/bin/genectl
```
For how to use genectl, you can see [genectl](https://kubegene.netlify.com/docs/guides/genectl-command/).

## Build your tool repo

A tool is a mirrored package of bioinformatics software that genome sequencing use. When you use genectl to submit your workflow, you should specify your tool repo address,it can be a directory or URL, if it is a URL, it must point to a specify tool file. The default tool repo is local directory “/${home}/kubegene/tools”). You can write your own tools and put them in the tool repo. For how to write your tool yaml, you can see [tool](https://kubegene.netlify.com/docs/guides/tool/).

## Write and submit your workflow.

We have defined a complete set of gene sequencing workflow grammars. It keeps the user’s traditional usage habit as much as possible and requires a very low learning cost to learn how to write and use the workflow. For how to write your workflow, you can see [workflow grammar](https://kubegene.netlify.com/docs/guides/workflow-grammar/).

We have several example workflow yaml in the example directory. A simple sample you can see in the `example/simple-sample` directory.

For submit the workflow, you need create the storage volume first:

```bash
$ kubectl create -f example/simale-sample/sample-pv.yaml
persistentvolume "sample-pv" created
$ kubectl create -f example/simale-sample/sample-pvc.yaml
persistentvolumeclaim "sample-pvc" created
```
Then you can submit the workflow by the following command:

```bash
$ genectl sub workflow example/simale-sample/simple-sample.yaml
The workflow has been submitted successfully! And the execution has been created.
Your can use the follow command to query the status of workflow execution.

        genectl get execution execution-bf0dc -n default

or use the follow command to query the detail info for workflow execution.

        genectl describe execution-bf0dc -n default

or use the follow command to delete the workflow execution.

        genectl del execution-bf0dc -n default
```

Query the detail execution status of workflow:

```
$ genectl describe execution-bf0dc -n default
Namespace:    default
Labels:       <none>
Annotations:  <none>
Phase:        Succeeded
Message:      execution has run successfully
workflow:
  jobprepare(total: 1; success: 1; failed: 0; running: 0; error: 0):
    Subtask                       Phase      Message
    -------                       -----      -------
    execution-bf0dc.jobprepare.0  Succeeded  success
  joba(total: 2; success: 2; failed: 0; running: 0; error: 0):
    Subtask                 Phase      Message
    -------                 -----      -------
    execution-bf0dc.joba.0  Succeeded  success
    execution-bf0dc.joba.1  Succeeded  success
  jobb(total: 3; success: 3; failed: 0; running: 0; error: 0):
    Subtask                 Phase      Message
    -------                 -----      -------
    execution-bf0dc.jobb.0  Succeeded  success
    execution-bf0dc.jobb.1  Succeeded  success
    execution-bf0dc.jobb.2  Succeeded  success
  jobc(total: 2; success: 2; failed: 0; running: 0; error: 0):
    Subtask                 Phase      Message
    -------                 -----      -------
    execution-bf0dc.jobc.1  Succeeded  success
    execution-bf0dc.jobc.0  Succeeded  success
  jobd(total: 2; success: 2; failed: 0; running: 0; error: 0):
    Subtask                 Phase      Message
    -------                 -----      -------
    execution-bf0dc.jobd.1  Succeeded  success
    execution-bf0dc.jobd.0  Succeeded  success
  jobfinish(total: 1; success: 1; failed: 0; running: 0; error: 0):
    Subtask                      Phase      Message
    -------                      -----      -------
    execution-bf0dc.jobfinish.0  Succeeded  success       
```

## Feature on the road  

KubeGene has provide the basic functionalities for running the genome sequencing workflow. More feature will be added:

- Support other workflow engine, such as argo and so on.
- User-defined workflow grammar plugin support.
- Support for hybrid cloud deployment.

## Support

If you need support, start with the [troubleshooting guide], and work your way through the process that we've outlined.

That said, if you have questions, reach out to us, feel free to reach out to these folks:

- @kevin-wangzefeng 
- @lichuqiang 
- @wackxu 
