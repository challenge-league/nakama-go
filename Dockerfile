FROM golang:1.13-alpine  as build-env

RUN mkdir /user && \
    echo 'nobody:x:65534:65534:nobody:/:' > /user/passwd && \
    echo 'nobody:x:65534:' > /user/group

ENV GO111MODULE=on
RUN apk --no-cache add make git gcc libtool musl-dev ca-certificates dumb-init && \
    rm -rf /var/cache/apk/* /tmp/*

WORKDIR /go/build/nakama-go

COPY --from=localhost:32000/nakama-apigrpc:dkozlov /go/build/nakama-apigrpc apigrpc
COPY context/go.mod context/go.mod
COPY context/go.sum context/go.sum
COPY commands/go.mod commands/go.mod
COPY commands/go.sum commands/go.sum
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN bash build.sh
#COPY ./go.mod ./go.sum ./
#RUN go mod download && rm go.mod go.sum
