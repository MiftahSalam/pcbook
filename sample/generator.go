package sample

import (
	"pc-book/pb"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewLaptop() *pb.Laptop {
	brand := randomLaptopBrand()
	return &pb.Laptop{
		Id:       uuid.New().String(),
		Brand:    brand,
		Name:     randomLaptopName(brand),
		Cpu:      NewCPU(),
		Memory:   NewRAM(),
		Gpus:     []*pb.GPU{NewGPU()},
		Storages: []*pb.Storage{NewSSD(), NewHDD()},
		Screen:   NewScreen(),
		Keyboard: NewKeyboard(),
		Weight: &pb.Laptop_WeightKg{
			WeightKg: randomFloat64(1.0, 3.0),
		},
		PriceUsd:    randomFloat64(1500, 3000),
		ReleaseYear: uint32(randomInt(2015, 2019)),
		UpdateAt:    timestamppb.Now(),
	}
}

func NewLaptopScore() float64 {
	return randomFloat64(1, 10)
}

func NewScreen() *pb.Screen {
	return &pb.Screen{
		Resolution: randomResolution(),
		SizeInch:   randomFloat32(13, 17),
		Panel:      randomScreen(),
		Multitouch: randomBool(),
	}
}

func NewGPU() *pb.GPU {
	brand := randomGpuBrand()
	minGHZ := randomFloat64(1.0, 1.5)

	return &pb.GPU{
		Brand:   brand,
		Name:    randomGPUName(brand),
		MinFreq: minGHZ,
		MaxFreq: randomFloat64(minGHZ, 5.0),
		Memory: &pb.Memory{
			Value: uint64(randomInt(2, 6)),
			Unit:  pb.Memory_GB,
		},
	}
}

func NewSSD() *pb.Storage {
	return &pb.Storage{
		Driver: pb.Storage_SSD,
		Memory: &pb.Memory{
			Value: uint64(randomInt(128, 1024)),
			Unit:  pb.Memory_GB,
		},
	}
}

func NewHDD() *pb.Storage {
	return &pb.Storage{
		Driver: pb.Storage_HDD,
		Memory: &pb.Memory{
			Value: uint64(randomInt(1, 6)),
			Unit:  pb.Memory_TB,
		},
	}
}

func NewRAM() *pb.Memory {
	return &pb.Memory{
		Value: uint64(randomInt(2, 6)),
		Unit:  pb.Memory_GB,
	}
}

func NewCPU() *pb.CPU {
	brand := randomCpuBrand()
	cores := randomInt(2, 8)
	minGHZ := randomFloat64(2.0, 3.5)
	return &pb.CPU{
		Brand:         brand,
		Name:          randomCpuName(brand),
		CoresMunber:   uint32(cores),
		ThreadsNumber: uint32(randomInt(cores, 12)),
		MinFreq:       minGHZ,
		MaxFreq:       randomFloat64(minGHZ, 5.0),
	}
}

func NewKeyboard() *pb.Keyboard {
	return &pb.Keyboard{
		Layout:  randomKeyboardLayout(),
		Backlit: randomBool(),
	}
}
