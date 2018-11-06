
.PHONY: all kube-dag genectl clean test e2ebin

IMAGE_NAME=kube-dag
TAG=$(shell git rev-parse --short HEAD)

REV=$(shell git describe --long --match='v*' --dirty)

ifdef V
TESTARGS = -v -args -alsologtostderr -v 10
else
TESTARGS =
endif


all: kube-dag genectl e2ebin

kube-dag:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-X main.version=$(REV) -extldflags "-static"' -o ./bin/kube-dag ./cmd/kube-dag

genectl:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-X main.version=$(REV) -extldflags "-static"' -o ./bin/genectl ./cmd/genectl

clean:
	-rm -rf bin

container: kube-dag
	docker build -t $(IMAGE_NAME):$(TAG) .

test:
	go test `go list ./... | grep -v -e 'vendor' -e 'test'` $(TESTARGS)
	go vet `go list ./... | grep -v vendor`
	./hack/e2e.sh
