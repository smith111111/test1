package utils

import (
	"crypto/rand"
	"errors"
	"log"
	"math/big"
	"time"

	"github.com/pborman/uuid"
)

func Random(max, min int) int {
	for {
		r, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
		if i := int(r.Int64()); i >= min {
			return i
		}
	}
}

func Percent(n, base int) bool {
	if n <= 0 {
		return false
	}
	if n > base {
		return false
	}
	var p = make([]int, n)
	for i := 0; i < n; i++ {
		p[i] = Random(base, 0)
	}
	d := Random(base+(base/2), 0)
	for i := 0; i < n; i++ {
		if d == p[i] {
			return true
		}
	}
	return false
}

func Guid() int64 {
	for i := 0; i < 3; i++ {
		t, _, err := uuid.GetTime()
		if err != nil {
			log.Println(err)
			continue
		}
		return int64(t)
	}
	log.Println(errors.New("newId generate error"))
	return time.Now().UnixNano()
}