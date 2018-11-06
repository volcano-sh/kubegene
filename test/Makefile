LDFLAGS :=-ldflags "-w -extldflags '-z relro -z now'"

all: gene2e

gene2e: Makefile
	go test $(LDFLAGS) -c -o gene2e

install: gene2e
	cp gene2e /usr/bin/

clean:
	rm -f gene2e

