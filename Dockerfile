FROM golang:1.14.3

WORKDIR /go/src/github.com/zkl94/pod-informer
COPY . .
RUN go build -ldflags "-linkmode external -extldflags -static" .

FROM scratch
COPY --from=0 /go/src/github.com/zkl94/pod-informer/pod-informer /pod-informer
CMD ["/pod-informer"]
