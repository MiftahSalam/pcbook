package service

import (
	"context"
	"pc-book/pb"
	"pc-book/sample"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateLaptopService(t *testing.T) {
	t.Parallel()

	laptopNoId := sample.NewLaptop()
	laptopNoId.Id = ""

	laptopInvalidId := sample.NewLaptop()
	laptopInvalidId.Id = "inalid"

	laptopDuplicateId := sample.NewLaptop()
	storeDulicateId := NewInMemoryLaptopStore()
	err := storeDulicateId.Save(laptopDuplicateId)
	require.Nil(t, err)

	testCases := []*struct {
		name   string
		laptop *pb.Laptop
		store  LaptopStore
		code   codes.Code
	}{
		{
			name:   "success_with_id",
			laptop: sample.NewLaptop(),
			store:  NewInMemoryLaptopStore(),
			code:   codes.OK,
		},
		{
			name:   "success_no_id",
			laptop: laptopNoId,
			store:  NewInMemoryLaptopStore(),
			code:   codes.OK,
		},
		{
			name:   "fail_invalid_id",
			laptop: laptopInvalidId,
			store:  NewInMemoryLaptopStore(),
			code:   codes.InvalidArgument,
		},
		{
			name:   "fail_duplicate_id",
			laptop: laptopDuplicateId,
			store:  NewInMemoryLaptopStore(),
			code:   codes.AlreadyExists,
		},
	}

	for i := 0; i < len(testCases); i++ {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := &pb.CreateLaptopRequest{
				Laptop: tc.laptop,
			}

			server := NewLaptopServer(tc.store, nil, nil)

			res, err := server.CreateLaptop(context.Background(), req)
			if tc.code == codes.OK {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.NotEmpty(t, res.Id)
				if len(tc.laptop.Id) > 0 {
					require.Equal(t, tc.laptop.Id, res.Id)
				} else {
					require.Error(t, err)
					require.Nil(t, res)
					st, ok := status.FromError(err)
					require.True(t, ok)
					require.Equal(t, tc.code, st)
				}
			}
		})
	}
}
