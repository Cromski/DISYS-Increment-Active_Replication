package main

import (
	"log"
	"main/increment"
	"net"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type frontend struct {
	increment.UnimplementedIncrementServiceServer
	replicas      map[string]increment.IncrementServiceClient
	replicasLock  sync.Mutex
	incrementLock sync.Mutex
}

func main() {
	log.Println(">> Loading FrontEnd now... Please wait!")

	lis, err := net.Listen("tcp", "localhost:9999")

	if err != nil {
		log.Fatalf("Failed to listen on port 9999, %s", err)
	}

	log.Print(">> Listener registered - setting up server now...")

	grpcServer := grpc.NewServer()

	fe := &frontend{replicas: make(map[string]increment.IncrementServiceClient)}
	fe.FindReplicas()
	go fe.Heartbeat()

	log.Print("===============================================================================")
	log.Print(" ")
	log.Print("                      Welcome to the Incrementor Service!                      ")
	log.Print("                                    FrontEnd                                   ")
	log.Print("                            Running on localhost:9999                          ")
	log.Print(" ")
	log.Print("===============================================================================")

	increment.RegisterIncrementServiceServer(grpcServer, fe)

	err = grpcServer.Serve(lis)

	if err != nil {
		log.Fatal("Failed to serve gRPC server over port 9999")
	}
}

func (fe *frontend) FindReplicas() {
	wg := sync.WaitGroup{}

	for i := 1000; i < 10000; i += 1000 {
		port := strconv.Itoa(i)

		if _, ok := fe.replicas[port]; ok {
			continue
		}

		wg.Add(1)

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			defer wg.Done()

			conn, err := grpc.DialContext(ctx, "localhost:"+port, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())

			if err != nil {
				return
			}

			log.Printf(">> Found server: localhost:%s", port)

			fe.replicasLock.Lock()
			fe.replicas[port] = increment.NewIncrementServiceClient(conn)
			fe.replicasLock.Unlock()
		}()
	}
	wg.Wait()
}

func (fe *frontend) Heartbeat() {
	for {
		fe.FindReplicas()
		time.Sleep(2 * time.Second)
	}
}

func (fe *frontend) Increment(ctx context.Context, req *increment.Request) (*increment.Value, error) {
	fe.incrementLock.Lock()
	defer fe.incrementLock.Unlock()

	var finalResponse *increment.Value
	log.Printf(">> Received Increment from client")

	wg := sync.WaitGroup{}

	for port, replica := range fe.replicas {
		wg.Add(1)

		go func(_port string, _replica increment.IncrementServiceClient) {
			defer wg.Done()

			response, err := _replica.Increment(context.Background(), req)

			if err != nil {
				fe.replicasLock.Lock()
				delete(fe.replicas, _port)
				fe.incrementLock.Unlock()
				return
			}

			finalResponse = response
		}(port, replica)
	}

	wg.Wait()

	return finalResponse, nil
}
