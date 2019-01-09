FROM golang:alpine as builder
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates
COPY . $GOPATH/src/aws-s3/
WORKDIR $GOPATH/src/aws-s3/
RUN go get -d -v
ARG CACHE_TAG
ARG SOURCE_COMMIT
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-X 'main.version=$CACHE_TAG' -X 'main.commit=$SOURCE_COMMIT'" -o /go/bin/aws-s3

FROM scratch
COPY --from=builder /go/bin/aws-s3 /go/bin/aws-s3
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/go/bin/aws-s3"]
