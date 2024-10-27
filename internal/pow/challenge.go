package pow

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"math/big"
	"math/rand"
	"time"
)

const (
	version1 = 1
	zero     = '0'
)

type Challenge interface {
	String() string
	Check() bool
	Compute(maxIterations uint64) error
	GetNonce() uint64
}

var (
	ErrMaxIterExceeded   = errors.New("max iterations amount exceeded")
	ErrFailedToUnmarshal = errors.New("failed to unmarshal challenge data")
)

type hashcash struct {
	Version  int
	Bits     int
	Date     time.Time
	Resource string
	Rand     []byte
	Counter  int64
}

func NewHashcash(bits int, resource string) Challenge {
	return &hashcash{
		Version:  version1,
		Bits:     bits,
		Date:     time.Now(),
		Resource: resource,
		Rand:     randBytes(),
	}
}

func (h *hashcash) String() string {
	return fmt.Sprintf("%d:%d:%s:%s:%s",
		h.Version,
		h.Bits,
		h.Resource,
		base64.StdEncoding.EncodeToString(h.Rand),
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", h.Counter))),
	)
}

func (h *hashcash) Check() bool {
	hashString := data2Sha1Hash(h.String())
	if h.Bits > len(hashString) {
		return false
	}

	for _, ch := range hashString[:h.Bits] {
		if ch != zero {
			return false
		}
	}
	return true
}

func (h *hashcash) Compute(maxIterations uint64) error {
	for h.Counter <= int64(maxIterations) {
		if h.Check() {
			return nil
		}
		h.Counter++
	}
	return ErrMaxIterExceeded
}

func (h *hashcash) GetNonce() uint64 {
	return binary.BigEndian.Uint64(h.Rand)
}

func Unmarshal(data []byte) (Challenge, error) {
	hash := &hashcash{}
	if err := msgpack.Unmarshal(data, &hash); err != nil {
		return nil, ErrFailedToUnmarshal
	}
	return hash, nil
}

func randBytes() []byte {
	return big.NewInt(int64(rand.Uint64())).Bytes()
}

func data2Sha1Hash(data string) string {
	h := sha1.New()
	h.Write([]byte(data))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
