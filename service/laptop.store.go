package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"pc-book/pb"
	"sync"

	"github.com/jinzhu/copier"
)

var ErrAlreadyExist = errors.New("record already exist")

type LaptopStore interface {
	Save(laptop *pb.Laptop) error
	Find(id string) (*pb.Laptop, error)
	Search(ctx context.Context, filter *pb.FilterMessage, found func(laptop *pb.Laptop) error) error
}

type InMemoryLaptopStore struct {
	mutex sync.RWMutex
	data  map[string]*pb.Laptop
}

func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

// Search implements LaptopStore.
func (store *InMemoryLaptopStore) Search(ctx context.Context, filter *pb.FilterMessage, found func(laptop *pb.Laptop) error) error {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	for _, laptop := range store.data {
		if ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded {
			log.Print("context is cancelled")

			return errors.New("context is cancelled")
		}

		if isQualified(filter, laptop) {
			other := &pb.Laptop{}
			err := copier.Copy(other, laptop)
			if err != nil {
				return fmt.Errorf("cannot copy laptop data: %w", err)
			}

			err = found(other)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.data[laptop.Id] != nil {
		return ErrAlreadyExist
	}

	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return fmt.Errorf("cannot copy laptop data: %w", err)
	}

	store.data[other.Id] = other

	return nil
}

func (store *InMemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	laptop := store.data[id]
	if laptop == nil {
		return nil, nil
	}

	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, fmt.Errorf("cannot copy laptop data: %w", err)
	}

	return other, nil
}

func isQualified(filter *pb.FilterMessage, laptop *pb.Laptop) bool {
	if laptop.GetPriceUsd() > filter.GetMaxPriceUsd() {
		return false
	}

	if laptop.GetCpu().GetCoresMunber() < filter.GetCpuCores() {
		return false
	}

	if laptop.GetCpu().GetMinFreq() < filter.GetMixCpuGhz() {
		return false
	}

	if toBit(laptop.GetMemory()) < toBit(filter.GetMinRam()) {
		return false
	}

	return true
}

func toBit(memory *pb.Memory) uint64 {
	switch memory.GetUnit() {
	case pb.Memory_BIT:
		return memory.GetValue()
	case pb.Memory_BYTE:
		return memory.GetValue() << 3
	case pb.Memory_KB:
		return memory.GetValue() << 13
	case pb.Memory_MB:
		return memory.GetValue() << 23
	case pb.Memory_GB:
		return memory.GetValue() << 33
	case pb.Memory_TB:
		return memory.GetValue() << 43
	default:
		return 0
	}
}
