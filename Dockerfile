FROM golang:1.17.2 as builder

WORKDIR $GOPATH/src/github.com/majeinfo/k8s-delete-protection
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN cd main && go test
RUN cd main && CGO_ENABLED=0 go build -o /go/bin/k8s-delete-protection

FROM scratch
COPY --from=builder /go/bin/k8s-delete-protection /go/bin/k8s-delete-protection
ENTRYPOINT ["/go/bin/k8s-delete-protection"]
CMD ["-cert", "/certs/server.pem", "-key", "/certs/server-key.pem"]
