package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestStorage(t *testing.T) {
	const keyTTL = 2 * time.Second

	storage := NewStorage(context.Background(), keyTTL)

	// 1. Add and get key
	key1 := uint64(1)
	storage.Add(key1)

	gotKey, err := storage.Get(key1)
	assert.NoError(t, err)
	assert.Equal(t, key1, gotKey)

	// 2. Get non existing key
	_, err = storage.Get(2)
	assert.Equal(t, ErrKeyNotFound, err)

	// 3. Delete key
	key2 := uint64(2)
	storage.Add(key2)

	gotKey, err = storage.Get(key2)
	assert.NoError(t, err)
	assert.Equal(t, key2, gotKey)

	storage.Delete(key2)

	_, err = storage.Get(key2)
	assert.Equal(t, ErrKeyNotFound, err)

	// 4. Wait till key would be cleaned up from db
	gotKey, err = storage.Get(key1)
	assert.NoError(t, err)
	assert.Equal(t, key1, gotKey)

	time.Sleep(keyTTL)

	_, err = storage.Get(key1)
	assert.Equal(t, ErrKeyNotFound, err)
}
