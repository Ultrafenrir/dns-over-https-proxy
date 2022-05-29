FROM golang:1.16.6-alpine3.14 as build-env
RUN mkdir -p $GOPATH/src/server/
WORKDIR $GOPATH/src/server/
COPY . .
RUN go mod init \
    && CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o proxy

FROM scratch
COPY --from=build-env /etc/ssl/cert.pem /etc/ssl/cert.pem
COPY --from=build-env /etc/passwd /etc/passwd
COPY --from=build-env /go/src/server/proxy .
EXPOSE 53
USER nobody
ENTRYPOINT ["/proxy"]
