FROM golang:1.16.2 as builder

WORKDIR $GOPATH/src/github.com/majeinfo/nodedisruptionbudget
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /go/bin/nodedisruptionbudget

FROM scratch
COPY --from=builder /go/bin/nodedisruptionbudget /go/bin/nodedisruptionbudget
ENTRYPOINT ["/go/bin/nodedisruptionbudget"]
CMD ["-cert", "/certs/server.pem", "-key", "/certs/server-key.pem"]
