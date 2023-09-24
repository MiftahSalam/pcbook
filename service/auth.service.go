package service

import (
	"context"
	"pc-book/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	userStore UserStore
	jwt       JwtManager
}

func NewAuthServer(userStore UserStore, jwtManger JwtManager) *AuthServer {
	return &AuthServer{
		userStore: userStore,
		jwt:       jwtManger,
	}
}

func (auth *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := auth.userStore.Find(req.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot find user: %v", err)
	}

	if user == nil || !user.IsCorrectPassword(req.GetPassword()) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid username/password")
	}

	token, err := auth.jwt.Generate(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create user token: %v", err)
	}

	return &pb.LoginResponse{
		AccessToken: token,
	}, nil
}
