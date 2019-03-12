
.PHONY: all kube-dag genectl clean test e2e

PACKAGE=kubegene.io/kubegene
CURRENT_DIR=$(shell pwd)
VERSION=$(shell git describe --long --match='v*' --dirty)
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_TAG=$(shell if [ -z "`git status --porcelain`" ]; then git describe --exact-match --tags HEAD 2>/dev/null; fi)
GIT_TREE_STATE=$(shell if [ -z "`git status --porcelain`" ]; then echo "clean" ; else echo "dirty"; fi)


override LDFLAGS += \
  -X ${PACKAGE}.version=${VERSION} \
  -X ${PACKAGE}.buildDate=${BUILD_DATE} \
  -X ${PACKAGE}.gitCommit=${GIT_COMMIT} \
  -X ${PACKAGE}.gitTreeState=${GIT_TREE_STATE}

ifneq (${GIT_TAG},)
override LDFLAGS += -X ${PACKAGE}.gitTag=${GIT_TAG}
endif

IMAGE_NAME=kube-dag
TAG=$(shell git rev-parse --short HEAD)

ifdef V
TESTARGS = -v -args -alsologtostderr -v 10
else
TESTARGS =
endif


all: kube-dag genectl

kube-dag:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '${LDFLAGS} -extldflags "-static"' -o ./bin/kube-dag ./cmd/kube-dag

genectl:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '${LDFLAGS} -extldflags "-static"' -o ./bin/genectl ./cmd/genectl

clean:
	-rm -rf bin

docker: kube-dag
	docker build -t $(IMAGE_NAME):$(TAG) .

test:
	go test `go list ./... | grep -v -e 'vendor' -e 'test'` $(TESTARGS)
	go vet `go list ./... | grep -v vendor`

e2e: KUBECONFIG?=$(HOME)/.kube/config

e2e:
	./hack/e2e.sh $(IMAGE_NAME):$(TAG) $(KUBECONFIG) ./bin/genectl
