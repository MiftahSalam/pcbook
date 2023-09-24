package client

import (
	"context"
	"pc-book/pb"
	"time"

	"google.golang.org/grpc"
)

type AuthClient struct {
	service            pb.AuthServiceClient
	username, password string
}

func NewAuthClient(cc grpc.ClientConnInterface, username, password string) *AuthClient {
	return &AuthClient{
		service:  pb.NewAuthServiceClient(cc),
		username: username,
		password: password,
	}
}

func (auth *AuthClient) Login() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := auth.service.Login(ctx, &pb.LoginRequest{
		Username: auth.username,
		Password: auth.password,
	})
	if err != nil {
		return "", err
	}

	return res.GetAccessToken(), nil
}
