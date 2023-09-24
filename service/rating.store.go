package service

import (
	"sync"
)

type Rating struct {
	Count uint32
	Sum   float64
}

type RatingStore interface {
	Add(laptopId string, score float64) (*Rating, error)
}

type InMemoryRatingStore struct {
	mutex  sync.RWMutex
	rating map[string]*Rating
}

func NewInMemoryRatingStore() RatingStore {
	return &InMemoryRatingStore{
		rating: make(map[string]*Rating),
	}
}

// Add implements RatingStore.
func (store *InMemoryRatingStore) Add(laptopId string, score float64) (*Rating, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	rating := store.rating[laptopId]
	if rating == nil {
		rating = &Rating{
			Count: 1,
			Sum:   score,
		}
	} else {
		rating.Count++
		rating.Sum += score
	}

	store.rating[laptopId] = rating

	return rating, nil
}
