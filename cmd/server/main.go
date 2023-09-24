package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"pc-book/pb"
	"pc-book/service"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	JWT_SECRET_KEY = "secret"
	JWT_DURATION   = 15 * time.Minute
)

func main() {
	port := flag.Int("port", 0, "server port")
	serverType := flag.String("server-type", "grpc", "type of server (grpc/rest)")
	flag.Parse()

	log.Printf("start server on port %d", *port)

	userStore := service.NewInMemoryUserStore()
	err := seedUser(userStore)
	if err != nil {
		log.Fatal("cannot seed users")
	}

	jwtManager := service.NewJwtManager(JWT_SECRET_KEY, JWT_DURATION)
	authServer := service.NewAuthServer(userStore, *jwtManager)
	authInterceptor := service.NewAuthInterceptor(jwtManager, accessableRoles())

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("tmp")
	ratingStore := service.NewInMemoryRatingStore()
	laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)

	if *serverType == "rest" {
		err = runRestServer(jwtManager, authServer, laptopServer, authInterceptor, port)
	} else {
		err = runGrpcServer(jwtManager, authServer, laptopServer, authInterceptor, port)
	}
	if err != nil {
		log.Fatal("cannot serve server: ", err)
	}
}

func runRestServer(
	jwtManager *service.JwtManager,
	authServer pb.AuthServiceServer,
	laptopServer pb.LaptopServiceServer,
	authInterceptor *service.AuthInterceptor,
	port *int,
) error {
	mux := runtime.NewServeMux()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := pb.RegisterAuthServiceHandlerServer(ctx, mux, authServer)
	if err != nil {
		return err
	}

	err = pb.RegisterLaptopServiceHandlerServer(ctx, mux, laptopServer)
	if err != nil {
		return err
	}
	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}

	log.Printf("starting rest server at %s", listener.Addr().String())

	return http.Serve(listener, mux)
}

func runGrpcServer(
	jwtManager *service.JwtManager,
	authServer pb.AuthServiceServer,
	laptopServer pb.LaptopServiceServer,
	authInterceptor *service.AuthInterceptor,
	port *int,
) error {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			authInterceptor.Unary(),
		),
		grpc.StreamInterceptor(
			authInterceptor.Stream(),
		),
	)

	pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	reflection.Register(grpcServer)

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}

	log.Printf("starting grpc server at %s", listener.Addr().String())
	return grpcServer.Serve(listener)
}

func accessableRoles() map[string][]string {
	const servicePath = "/LaptopService/"

	return map[string][]string{
		servicePath + "CreateLaptop": {"admin"},
		servicePath + "UploadImage":  {"admin"},
		servicePath + "RateLaptop":   {"admin", "user"},
	}
}

func seedUser(store service.UserStore) error {
	err := createUser(store, "admin", "secret", "admin")
	if err != nil {
		return err
	}

	return createUser(store, "user1", "secret", "user")
}

func createUser(store service.UserStore, username, password, role string) error {
	user, err := service.NewUser(username, password, role)
	if err != nil {
		return err
	}

	return store.Save(user)
}
