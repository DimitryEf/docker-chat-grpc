//https://github.com/protocolbuffers/protobuf/releases/tag/v3.10.0		protoc
//go get -u google.golang.org/grpc
//go get -u github.com/golang/protobuf/protoc-gen-go
//protoc -I proto --go_out=plugins=grpc:proto proto/service.proto

//docker build --tag=docker-chat-grpc .
//docker run -it -p 8080:8080 docker-chat-grpc

package main

import (
	"github.com/DimitryEf/docker-chat-grpc/proto"
	"log"
	"net"
	"os"
	"sync"

	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
)

var grpcLog glog.LoggerV2

func init() {
	grpcLog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

type Connection struct {
	stream proto.Broadcast_CreateStreamServer
	id     string
	active bool
	error  chan error
}

type Server struct {
	Connection []*Connection
}

func (s *Server) CreateStream(pconn *proto.Connect, stream proto.Broadcast_CreateStreamServer) error {
	conn := &Connection{
		stream: stream,
		id:     pconn.User.Id,
		active: true,
		error:  make(chan error),
	}

	s.Connection = append(s.Connection, conn)

	return <-conn.error
}

func (s *Server) BroadcastMessage(ctx context.Context, msg *proto.Message) (*proto.Close, error) {
	wait := sync.WaitGroup{}
	done := make(chan int)

	for _, conn := range s.Connection {
		wait.Add(1)

		go func(msg *proto.Message, conn *Connection) {
			defer wait.Done()

			if conn.active {
				err := conn.stream.Send(msg)
				grpcLog.Info("Sending message to: ", conn.stream)

				if err != nil {
					grpcLog.Errorf("Error with Stream: %v - Error: %v", conn.stream, err)
					conn.active = false
					conn.error <- err
				}
			}
		}(msg, conn)

	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
	return &proto.Close{}, nil
}

func main() {
	var connections []*Connection

	server := &Server{connections}

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("error creating the server %v", err)
	}

	grpcLog.Info("Starting server at port :8080")

	proto.RegisterBroadcastServer(grpcServer, server)
	grpcServer.Serve(listener)
}
