package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	port := 50051
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("gRPC server listening at %v", lis.Addr())

	select {}
}
