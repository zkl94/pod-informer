FROM golang:latest

WORKDIR /go/src/github.com/zkl94/pod-informer
COPY . .
RUN go build -ldflags "-linkmode external -extldflags -static" -a main.go

FROM scratch
COPY --from=0 /go/src/github.com/purplebooth/example/main /main
CMD ["/main"]
