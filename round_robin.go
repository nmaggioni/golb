package golb

import (
	"fmt"
	"net"
	"sync"
)

type rrIndex struct {
	sync.Mutex
	i int
}

var index rrIndex

func incrementRRIndex() {
	index.Lock()
	defer index.Unlock()

	if index.i < len(config.Upstreams)-1 {
		index.i++
	} else {
		index.i = 0
	}
}

func roundRobin(id string, conn net.Conn, tries int) {
	if tries < len(config.Upstreams)*config.MaxCycles {
		err := forward(id, conn, config.Upstreams[index.i])
		incrementRRIndex()
		if err != nil {
			roundRobin(id, conn, tries+1)
		}
	} else {
		if config.Verbose {
			fmt.Printf("%s - Max retry cycles reached, aborting\n", id)
		}
		conn.Close()
	}
}
