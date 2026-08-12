package evaluation

import (
	"hash/fnv"
	"math/rand"
	"strconv"
)

func stickinessHash(flagKey, stickinessValue string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(flagKey + ":" + stickinessValue))
	return h.Sum32() % 10000
}

func randomStickiness() string {
	return strconv.Itoa(rand.Intn(1_000_000_000))
}
