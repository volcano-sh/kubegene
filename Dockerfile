FROM alpine

COPY ./bin/kube-dag /usr/local/bin
ENTRYPOINT ["/usr/local/bin/kube-dag"]