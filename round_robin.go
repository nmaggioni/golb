package golb

import (
	"net"
	"sync"
)

type rr_index struct {
	sync.Mutex
	i int
}

var index rr_index

func incrementRRIndex() {
	index.Lock()
	defer index.Unlock()

	if index.i < len(config.Upstreams)-1 {
		index.i++
	} else {
		index.i = 0
	}
}

func roundRobin(id string, conn net.Conn) {
	err := forward(id, conn, config.Upstreams[index.i])
	incrementRRIndex()
	if err != nil {
		roundRobin(id, conn)
	}
}
