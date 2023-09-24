package service

import (
	"context"
	"net"
	"pc-book/pb"
	"pc-book/sample"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestLaptopClient(t *testing.T) {
	t.Parallel()

	laptopServer, serverAddr := startLaptopServer(t)
	laptopClient := newLaptopCient(t, serverAddr)
	laptop := sample.NewLaptop()
	expectedId := laptop.Id
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	res, err := laptopClient.CreateLaptop(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, expectedId, res.Id)

	other, err := laptopServer.laptopStore.Find(res.Id)
	require.NoError(t, err)
	require.NotNil(t, other)

	requireSameLaptop(t, laptop, other)
}

func startLaptopServer(t *testing.T) (*LaptopServer, string) {
	laptop := NewLaptopServer(NewInMemoryLaptopStore(), NewDiskImageStore(""), NewInMemoryRatingStore())
	grpcServer := grpc.NewServer()

	pb.RegisterLaptopServiceServer(grpcServer, laptop)

	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go grpcServer.Serve(listener)

	return laptop, listener.Addr().String()
}

func newLaptopCient(t *testing.T, serverAddr string) pb.LaptopServiceClient {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	return pb.NewLaptopServiceClient(conn)
}

func requireSameLaptop(t *testing.T, laptop1, laptop2 *pb.Laptop) {
	json1, err := protojson.Marshal(laptop1)
	require.NoError(t, err)

	json2, err := protojson.Marshal(laptop2)
	require.NoError(t, err)

	require.Equal(t, json1, json2)
}
