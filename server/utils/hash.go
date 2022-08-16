package utils

import (
	"hash"
	"io"
	"strconv"
)

func HashString(h hash.Hash, s string) {
	_, _ = io.WriteString(h, s)
}

func HashUint64(h hash.Hash, i uint64) {
	_, _ = io.WriteString(h, strconv.FormatUint(i, 10))
}

func HashInt(h hash.Hash, i int) {
	_, _ = io.WriteString(h, strconv.Itoa(i))
}
