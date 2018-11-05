## build client

bash hack/build-gen.sh && bash hack/update-all.sh -v

## build binary

bash build/build_all.sh 

## build release

bash build/release.sh

## CRD files

> $ ls pkg/apis/crd
>
>  components.gene.kubedag.yaml  executions.gene.kubedag.yaml
