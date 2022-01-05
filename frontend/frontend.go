package main

import (
	"flag"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"main/increment"
	"net"
	"strconv"
	"sync"
	"time"
)

type frontend struct{
	increment.UnimplementedIncrementServiceServer
	replicas map[string]increment.IncrementServiceClient
}

func main (){
	var myPort = flag.String("port", "", "")
	flag.Parse()

	log.Printf("Listening on port %v", *myPort)
	lis, err := net.Listen("tcp", "localhost:" + *myPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %v, %v", myPort, err)
	}

	grpcServer := grpc.NewServer()

	fe := &frontend{replicas: make(map[string]increment.IncrementServiceClient)}

	fe.FindReplicas()

	log.Println("we ready :)")

	go fe.Heartbeat()

	increment.RegisterIncrementServiceServer(grpcServer, fe)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to server %v", err)
	}
}

func (fe *frontend) Connect(port int) {
	conn, err := grpc.Dial("localhost:" + strconv.Itoa(port), grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(2*time.Second))

	if err != nil {
		return
	}
	fe.replicas[strconv.Itoa(port)] = increment.NewIncrementServiceClient(conn)
	log.Printf("Found this server: %i", strconv.Itoa(port))
}

func (fe *frontend) FindReplicas () {

	wg := sync.WaitGroup{}

	for i :=  1000; i < 10000; i+=1000 {
		if _, ok := fe.replicas[strconv.Itoa(i)]; ok {
			continue
		}
		wg.Add(1)
		go func (port int) {
			defer wg.Done()
			conn, err := grpc.Dial("localhost:" + strconv.Itoa(port), grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(2*time.Second))

			if err != nil {
				return
			}
			fe.replicas[strconv.Itoa(port)] = increment.NewIncrementServiceClient(conn)
			log.Printf("Found this server: %i", strconv.Itoa(port))
		} (i)
	}
	wg.Wait()
}

func (fe *frontend) Heartbeat (){
	for {
		fe.FindReplicas()
		time.Sleep(2*time.Second)
	}
}

func (fe *frontend) Increment(ctx context.Context, req *increment.Request) (*increment.Value, error){

	var finalResponse *increment.Value

	log.Printf("received request %s", req)

	for port, replica := range fe.replicas{
		response, err := replica.Increment(context.Background(), req)
		if err != nil {
			delete(fe.replicas, port)
		}
		finalResponse = response
	}

	return finalResponse, nil
}