FROM golang:alpine as build-env

ENV GO111MODULE=on

RUN apk update && apk add bash ca-certificates git gcc g++ libc-dev

RUN mkdir /docker-chat-grpc
RUN mkdir -p /docker-chat-grpc/proto

WORKDIR /docker-chat-grpc

COPY ./proto/service.pb.go /docker-chat-grpc/proto
COPY ./main.go /docker-chat-grpc

COPY go.mod .
COPY go.sum .

RUN go mod download

RUN go build -o docker-chat-grpc .

CMD ./docker-chat-grpc