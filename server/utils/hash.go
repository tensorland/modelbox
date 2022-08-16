package utils

import (
	"hash"
	"io"
	"strconv"

	"google.golang.org/protobuf/types/known/structpb"
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

func HashMeta(h hash.Hash, m map[string]*structpb.Value) {
	for k, v := range m {
		HashString(h, k)
		HashString(h, v.String())
	}
}
