package utils

import (
	"crypto/sha1"
	"fmt"
)

func Data2Sha1Hash(data string) string {
	h := sha1.New()
	h.Write([]byte(data))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
