package main

import (
	"flag"
	"log"
	"main/increment"
	"net"
	"os"
	"strconv"
	"sync"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
)

type server struct {
	increment.UnimplementedIncrementServiceServer
	value int32
	lock  sync.Mutex
}

func main() {
	var myPort = flag.String("port", "", "")
	flag.Parse()

	log.Printf("Listening on port %v", *myPort)
	lis, err := net.Listen("tcp", "localhost:"+*myPort)

	if err != nil {
		log.Fatalf("Failed to listen on port %v, %v", myPort, err)
	}

	grpcServer := grpc.NewServer()

	s := &server{value: 0}

	data, _ := os.ReadFile("number.txt")
	value, _ := strconv.ParseInt(string(data), 10, 64)
	s.value = int32(value)

	log.Printf("Loaded current increment value from file: %v", s.value)

	increment.RegisterIncrementServiceServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to server %v", err)
	}
}

func (s *server) Increment(ctx context.Context, req *increment.Request) (*increment.Value, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Printf("received request %s", req)

	value := increment.Value{
		Value: s.value,
	}

	s.value++

	go func() {
		os.WriteFile("./number.txt", []byte(strconv.Itoa(int(s.value))), 0644)
	}()

	return &value, nil
}
