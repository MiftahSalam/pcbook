package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"pc-book/pb"
	"pc-book/sample"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRateLaptop(client pb.LaptopServiceClient) {
	n := 3
	laptopIds := make([]string, n)

	for i := 0; i < n; i++ {
		laptop := sample.NewLaptop()
		laptopIds[i] = laptop.GetId()
		CreateLaptop(client, laptop)
	}

	scores := make([]float64, n)
	for {
		fmt.Print("rate laptop (y/n)?")
		var answer string
		fmt.Scan(&answer)

		if strings.ToLower(answer) != "y" {
			break
		}

		for i := 0; i < n; i++ {
			scores[i] = sample.NewLaptopScore()
		}

		err := RateLaptops(client, laptopIds, scores)
		if err != nil {
			log.Fatal(err)
		}
	}
}
func RateLaptops(client pb.LaptopServiceClient, laptopIds []string, scores []float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.RateLaptop(ctx)
	if err != nil {
		return fmt.Errorf("cannot rate laptop: %v", err)
	}

	waitRespone := make(chan error)
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				log.Print("no more receved data")
				waitRespone <- nil
				return
			}
			if err != nil {
				waitRespone <- fmt.Errorf("cannot receive stream response: %v", err)
				return
			}

			log.Print("received reaponse: ", res)
		}
	}()

	for i, laptopId := range laptopIds {
		err := stream.Send(&pb.RateLaptopRequest{
			LaptopId: laptopId,
			Score:    scores[i],
		})
		if err != nil {
			return fmt.Errorf("cannot send stream request: %v", err)
		}

		log.Print("sending request...")
	}

	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("cannot close send: %v", err)
	}

	err = <-waitRespone

	return err
}
func UploadImage(client pb.LaptopServiceClient, laptopId string, imagePath string) {
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatalf("cannot open image file: %v", err)
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.UploadImage(ctx)
	if err != nil {
		log.Fatalf("cannot upload image: %v", err)
	}

	err = stream.Send(&pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptopId,
				ImageType: filepath.Ext(imagePath),
			},
		},
	})
	if err != nil {
		log.Fatalf("cannot send image file: %v", err)
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("cannot read chunk to buffer: %v", err)
		}

		err = stream.Send(&pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		})
		if err != nil {
			log.Fatalf("cannot send chunk to server: %v", err)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		err2 := stream.RecvMsg(nil)
		log.Fatalf("cannot receive response from server: %v, %v", err, err2)
	}

	log.Printf("image uploaded with id %s, size %d", res.GetId(), res.GetSize())
}

func TestUploadImage(client pb.LaptopServiceClient) {
	laptop := sample.NewLaptop()
	CreateLaptop(client, laptop)
	UploadImage(client, laptop.GetId(), "./img/laptop.jpg")
}

func SearchLaptop(client pb.LaptopServiceClient, filter *pb.FilterMessage) {
	log.Print("search filter: ", filter)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.SearchLaptopRequest{
		Filter: filter,
	}
	strea, err := client.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatal("cannot search laptop", err)
	}

	for {
		res, err := strea.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatal("cannot receive response", err)
		}

		laptop := res.GetLaptop()

		log.Print("-- found: ", laptop.GetId())
		log.Print("  + brand: ", laptop.GetBrand())
		log.Print("  + name: ", laptop.GetName())
		log.Print("  + cpu core: ", laptop.GetCpu().GetCoresMunber())
		log.Print("  + cpu min freq: ", laptop.Cpu.GetMinFreq())
		log.Print("  + ram: ", laptop.GetMemory().GetValue())
		log.Print("  + price: ", laptop.GetPriceUsd())
	}
}

func CreateLaptop(client pb.LaptopServiceClient, laptop *pb.Laptop) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	res, err := client.CreateLaptop(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Print("laptop already exist")
		} else {
			log.Fatal("cannot create laptop: ", err)
		}

		return
	}

	log.Printf("created laptop with id: %s", res.Id)
}
