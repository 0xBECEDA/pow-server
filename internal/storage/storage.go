package storage

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

type Store interface {
	Add(key uint64)
	Get(key uint64) (uint64, error)
	Delete(key uint64)
}

var ErrKeyNotFound = errors.New("key does not exist in storage")

type inMemoryDB struct {
	memoryDB map[uint64]time.Time
	keyTTL   time.Duration
	rw       *sync.RWMutex
}

func NewStorage(ctx context.Context, keyTTL time.Duration) Store {
	db := &inMemoryDB{
		memoryDB: make(map[uint64]time.Time),
		rw:       &sync.RWMutex{},
		keyTTL:   keyTTL,
	}

	go db.clean(ctx) // clean up old challenges
	return db
}

func (r *inMemoryDB) Add(key uint64) {
	r.rw.Lock()
	defer r.rw.Unlock()

	r.memoryDB[key] = time.Now().Add(r.keyTTL)
	log.Printf("added key: %d", key)
}

func (r *inMemoryDB) Get(key uint64) (uint64, error) {
	log.Printf("getting key: %d", key)

	r.rw.RLock()
	defer r.rw.RUnlock()

	_, ok := r.memoryDB[key]
	if ok {
		return key, nil
	}

	return 0, ErrKeyNotFound
}

func (r *inMemoryDB) Delete(key uint64) {
	log.Printf("deleting key: %d", key)

	r.rw.Lock()
	defer r.rw.Unlock()

	delete(r.memoryDB, key)
}

// clean removes expired keys.
func (r *inMemoryDB) clean(ctx context.Context) {
	tick := time.NewTicker(r.keyTTL)

	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			now := time.Now()
			log.Printf("clean up storage started at %v", now)

			r.rw.Lock()
			for key, ttl := range r.memoryDB {
				if ttl.Before(now) {
					delete(r.memoryDB, key)
					log.Printf("key %v expired and was deleted", key)
				}
			}
			r.rw.Unlock()
		}
	}
}
