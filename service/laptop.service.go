package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"pc-book/pb"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const MAX_IMAGE_SIZE = 1 << 20

type LaptopServer struct {
	pb.UnimplementedLaptopServiceServer
	laptopStore LaptopStore
	imageStore  ImageStore
	ratingStore RatingStore
}

func NewLaptopServer(store LaptopStore, imgStore ImageStore, ratingStore RatingStore) *LaptopServer {
	return &LaptopServer{laptopStore: store, imageStore: imgStore, ratingStore: ratingStore}
}

func (server *LaptopServer) RateLaptop(stream pb.LaptopService_RateLaptopServer) error {
	for {
		err := contexError(stream.Context())
		if err != nil {
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			log.Fatalf("cannot receive stream request: %v", err)
			return status.Errorf(codes.Internal, "annot receive stream request")
		}

		log.Printf("received a rate-laptop request id %s, score %.2f", req.GetLaptopId(), req.GetScore())

		laptop, err := server.laptopStore.Find(req.GetLaptopId())
		if err != nil {
			log.Fatalf("cannot find laptop: %v", err)
			return status.Errorf(codes.Internal, "cannot find laptop")
		}
		if laptop == nil {
			log.Fatalf("laptop does not exist: %v", err)
			return status.Errorf(codes.NotFound, "laptop does not exist")
		}

		rating, err := server.ratingStore.Add(laptop.Id, req.GetScore())
		if err != nil {
			log.Fatalf("cannot add rating laptop: %v", err)
			return status.Errorf(codes.Internal, "cannot add rating laptop")
		}

		err = stream.Send(&pb.RateLaptopResponse{
			LaptopId:     laptop.GetId(),
			RateCount:    rating.Count,
			AverageScore: rating.Sum / float64(rating.Count),
		})
		if err != nil {
			log.Fatalf("cannot send rating response: %v", err)
			return status.Errorf(codes.Internal, "cannot send rating response")
		}
	}

	return nil
}

func (server *LaptopServer) UploadImage(stream pb.LaptopService_UploadImageServer) error {
	req, err := stream.Recv()
	if err != nil {
		log.Fatalf("cannot receive image info: %v", err)
		return status.Errorf(codes.Internal, "cannot receive image info")
	}

	laptopId, imageType := req.GetInfo().LaptopId, req.GetInfo().GetImageType()

	log.Printf("receive and upload-image request for laptop %s with image type %s", laptopId, imageType)

	laptop, err := server.laptopStore.Find(laptopId)
	if err != nil {
		log.Fatalf("cannot find laptop: %v", err)
		return status.Errorf(codes.Internal, "cannot find laptop")
	}
	if laptop == nil {
		log.Fatalf("laptop does not exist: %v", err)
		return status.Errorf(codes.NotFound, "laptop does not exist")
	}

	imageData := bytes.Buffer{}
	imageSize := 0

	for {
		log.Print("waiting to receive more data...")

		err := contexError(stream.Context())
		if err != nil {
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			log.Fatalf("cannot receive chunk data: %v", err)
			return status.Errorf(codes.Unknown, "cannot receive chunk data")
		}

		chunk := req.GetChunkData()
		size := len(chunk)

		imageSize += size
		if imageSize > MAX_IMAGE_SIZE {
			log.Fatal("image data too large")
			return status.Errorf(codes.InvalidArgument, "image data too large")
		}

		_, err = imageData.Write(chunk)
		if err != nil {
			log.Fatalf("cannot write chunk data: %v", err)
			return status.Errorf(codes.Internal, "cannot write chunk data")
		}
	}

	imageId, err := server.imageStore.Save(laptopId, imageType, imageData)
	if err != nil {
		log.Fatalf("cannot save image to the store: %v", err)
		return status.Errorf(codes.Internal, "cannot save image to the store")
	}

	err = stream.SendAndClose(&pb.UploadImageResponse{
		Id:   imageId,
		Size: uint32(imageSize),
	})
	if err != nil {
		log.Fatalf("cannot send response: %v", err)
		return status.Errorf(codes.Unknown, "cannot send response")
	}

	log.Printf("saved image with id %s and size %d", imageId, imageSize)

	return nil
}

func (server *LaptopServer) SearchLaptop(req *pb.SearchLaptopRequest, stream pb.LaptopService_SearchLaptopServer) error {
	filter := req.GetFilter()

	log.Printf("receive search-filter request with filter: %v", filter)

	err := server.laptopStore.Search(stream.Context(), filter, func(laptop *pb.Laptop) error {
		res := &pb.SearchLaptopResponse{Laptop: laptop}

		err := stream.Send(res)
		if err != nil {
			return err
		}

		log.Printf("send laptop with id: %s", laptop.GetId())

		return nil
	})
	if err != nil {
		return status.Errorf(codes.Internal, "unexpected error: %v", err)
	}

	return nil
}

func (server *LaptopServer) CreateLaptop(ctx context.Context, req *pb.CreateLaptopRequest) (*pb.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()

	log.Printf("receive a create laptop request with id: %s", laptop.Id)

	if len(laptop.Id) > 0 {
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "laptop id is not valid uuid: %v", err)
		}
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot generate laptop id =: %v", err)
		}

		laptop.Id = id.String()
	}

	err := contexError(ctx)
	if err != nil {
		return nil, err
	}

	err = server.laptopStore.Save(laptop)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, ErrAlreadyExist) {
			code = codes.AlreadyExists
		}

		return nil, status.Errorf(code, "cannot save laptop to the store: %v", err)
	}

	log.Printf("saved laptop to the store with id: %s", laptop.Id)

	return &pb.CreateLaptopResponse{
		Id: laptop.Id,
	}, nil
}

func contexError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		log.Print("deadline is canceled")
		return status.Errorf(codes.Canceled, "deadline is canceled")
	case context.DeadlineExceeded:
		log.Print("deadline is exceeded")
		return status.Errorf(codes.DeadlineExceeded, "deadline is exceed")
	default:
		return nil
	}
}
