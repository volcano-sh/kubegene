# KubeGene
[![CircleCI](https://circleci.com/gh/Huawei-PaaS/kubegene.svg?style=svg)](https://circleci.com/gh/Huawei-PaaS/kubegene)

## What is KubeGene
The KubeGene is dedicated to making genome sequencing process simple, portable and scalable. It provides a complete solution for genome sequencing on the kubernetes and supports mainstream biological genome sequencing scenarios such as DNA, RNA, and liquid biopsy. KubeGene is based on lightweight container technology and official standard algorithms. You can make a flexible and customizable genome sequencing process by using KubeGene.

KubeGene which running on the kubernetes makes genome sequencing simple and easy. It has the following characteristics:

+ **Universal workflow design grammar.** KubeGene provides a complete set of genome sequencing workflow grammars which decouples with specific analysis tools. It requires a very low learning cost to learn how to write and use the workflow. You can easily migrate the genome sequencing business to KubeGene.

+ **Tailor-made workflow for the biosequencing industry.** The workflow grammar is designed by comparing different genome sequencing scenarios. It also keeps the user’s traditional usage habit as much as possible and is closer to user scenarios.

+ **More efficient resource usage.** KubeGene uses container to run the genome sequencing business. Compared to traditional genome sequencing solutions using virtual machines, KubeGene makes resource usage more efficient and avoiding resource idle.

+ **Scaling based on demand.** Kubernetes can automatically scale your cluster based on your genome sequencing workload by using KubeGene. Also you can easily scale the kubernetes cluster manually. It can save your production costs. 


## Devel

### build client
```
bash hack/build-gen.sh && bash hack/update-all.sh -v
```
### build binary
```
bash build/build_all.sh 
```
### build release
```
bash build/release.sh
```
### CRD files
```
> $ ls pkg/apis/crd
>
>  components.gene.kubedag.yaml  executions.gene.kubedag.yaml
```
