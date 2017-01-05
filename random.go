package golb

import (
	"math/rand"
	"net"
	"time"
)

func random(id string, conn net.Conn, tries int) {
	seed := rand.NewSource(time.Now().UnixNano())
	randSeeded := rand.New(seed)
	i := randSeeded.Intn(len(config.Upstreams))

	err := forward(id, conn, config.Upstreams[i])
	incrementRRIndex()
	if err != nil {
		forwardWithStrategy(id, conn, tries+1)
	}
}
