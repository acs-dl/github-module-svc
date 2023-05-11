FROM golang:1.19-alpine as buildbase

RUN apk add git build-base

WORKDIR /go/src/github.com/acs-dl/github-module-svc
COPY vendor .
COPY . .

RUN GOOS=linux go build  -o /usr/local/bin/github-module /go/src/github.com/acs-dl/github-module-svc


FROM alpine:3.9

COPY --from=buildbase /usr/local/bin/github-module /usr/local/bin/github-module
RUN apk add --no-cache ca-certificates

ENTRYPOINT ["github-module"]
