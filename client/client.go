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
		log.Fatalf("Frontend prolly not running")
	}

	client := increment.NewIncrementServiceClient(conn)

	log.Println("yes yes :) type inc to inc :)")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "inc"{
			value, err := client.Increment(context.Background(), &increment.Request{})
			if err != nil {
				log.Fatalf("yep :)")
			}
			log.Println("previous value: " + strconv.Itoa(int(value.Value)))

		} else {
			log.Println("you can only type inc mfka")
		}
	}

}
