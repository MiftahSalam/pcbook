package main

import (
	"flag"
	"log"
	lClient "pc-book/client"
	"pc-book/pb"
	"pc-book/sample"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	username        = "admin"
	password        = "secret"
	refreshDuration = 30 * time.Second
)

func main() {
	serverAddr := flag.String("address", "0.0.0.0:8080", "server address")
	flag.Parse()

	log.Printf("dial server: %s", *serverAddr)

	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("cannot dial server: %v", err)
	}

	auth := lClient.NewAuthClient(conn, username, password)
	interceptor, err := lClient.NewAuthInterceptor(auth, accessableRoles(), refreshDuration)
	if err != nil {
		log.Fatalf("cannot create interceptor: %v", err)
	}
	serverOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
		grpc.WithUnaryInterceptor(
			interceptor.Unary(),
		),
		grpc.WithStreamInterceptor(
			interceptor.Stream(),
		),
	}
	conn2, err := grpc.Dial(
		*serverAddr,
		serverOptions...,
	)
	if err != nil {
		log.Fatalf("cannot dial server: %v", err)
	}

	laptopClient := pb.NewLaptopServiceClient(conn2)

	for i := 0; i < 10; i++ {
		lClient.CreateLaptop(laptopClient, sample.NewLaptop())
	}

	lClient.SearchLaptop(laptopClient, &pb.FilterMessage{
		MaxPriceUsd: 3000,
		CpuCores:    2,
		MixCpuGhz:   1.5,
		MinRam: &pb.Memory{
			Value: 2,
			Unit:  pb.Memory_GB,
		},
	})

	// testUploadImage(laptopClient)
	lClient.TestRateLaptop(laptopClient)
}

func accessableRoles() map[string]bool {
	const servicePath = "/LaptopService/"

	return map[string]bool{
		servicePath + "CreateLaptop": true,
		servicePath + "UploadImage":  true,
		servicePath + "RateLaptop":   true,
	}
}
