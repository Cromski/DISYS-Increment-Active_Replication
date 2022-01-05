package main

import (
	"bufio"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"main/increment"
	"os"
	"strconv"
	"time"
)

func main(){

	conn, err := grpc.Dial("localhost:10000", grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(2*time.Second))

	if err != nil {
		log.Fatalf("Frontend probably not running \nNOTE: Check if frontend port is correct")
	}

	client := increment.NewIncrementServiceClient(conn)

	log.Println("Successfully connected to our increment program \nType 'inc' to increment the counter:")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "inc"{
			value, err := client.Increment(context.Background(), &increment.Request{})
			if err != nil {
				log.Fatalf("Couldn't increase the counter")
			}
			log.Println("previous value: " + strconv.Itoa(int(value.Value)))

		} else {
			log.Println("You have to enter 'inc' to increase the value")
		}
	}

}
