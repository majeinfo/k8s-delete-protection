FROM golang:1.17.2 as builder

WORKDIR $GOPATH/src/github.com/majeinfo/nodedisruptionbudget
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN cd main && CGO_ENABLED=0 go build -o /go/bin/nodedisruptionbudget/main

FROM scratch
COPY --from=builder /go/bin/nodedisruptionbudget /go/bin/nodedisruptionbudget
ENTRYPOINT ["/go/bin/nodedisruptionbudget"]
CMD ["-cert", "/certs/server.pem", "-key", "/certs/server-key.pem"]
