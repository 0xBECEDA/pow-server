package pow

import (
	"world-of-wisdom/internal/storage"
)

type Service interface {
	Add(indicator uint64)
	Exists(indicator uint64) bool
	Delete(indicator uint64)
}

type hashCashService struct {
	storage storage.Store
}

func NewChallengeService(storage storage.Store) Service {
	return &hashCashService{
		storage: storage,
	}
}

func (repo *hashCashService) Add(indicator uint64) {
	repo.storage.Add(indicator)
}

func (repo *hashCashService) Exists(indicator uint64) bool {
	_, err := repo.storage.Get(indicator)
	if err != nil {
		return false
	}
	return true
}

func (repo *hashCashService) Delete(indicator uint64) {
	repo.storage.Delete(indicator)
}
