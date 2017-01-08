package golb

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/satori/go.uuid"
)

func Listen() error {
	if config.Verbose {
		fmt.Printf("LOAD - - Loaded configuration from file: %s\n", confPath)
	}

	listener, err := net.Listen("tcp", config.IP+":"+config.Port)
	if err != nil {
		return fmt.Errorf("Failed to setup listener: %v", err)
	}
	host := config.IP
	if host == "" {
		host = "127.0.0.1"
	}

	if config.Sticky {
		if config.Verbose {
			fmt.Println("LOAD - - Kickstarting GC...")
		}
		go staleSessionsCleaner()
	}

	fmt.Printf("LOAD - - Listening on %s:%s...\n", host, config.Port)

	for {
		conn, err := listener.Accept()
		id := uuid.NewV1().String()
		if err != nil || conn == nil {
			fmt.Fprintf(os.Stderr, "ERR  - - Failed to accept listener: %v\n", err)
			continue
		}

		if config.Verbose {
			fmt.Printf("INFO - %s - Accepted connection from %s\n", id, conn.RemoteAddr().String())
		}

		go func() {
			if config.Sticky {
				if session, isCached := getSession(stripPortFromAddr(conn.RemoteAddr().String())); isCached {
					if config.Verbose {
						fmt.Printf("INFO - %s - Client is sticked to %s\n", id, session.Upstream.Name)
					}
					err := forward(id, conn, session.Upstream)
					if err == nil {
						return
					}
				}
			}
			forwardWithStrategy(id, conn, 0)
		}()
	}
}

func forwardWithStrategy(id string, conn net.Conn, tries int) {
	if tries < len(config.Upstreams)*config.MaxCycles {
		switch config.Strategy {
		case "random":
			random(id, conn, tries)
		case "round-robin", "weighted-round-robin":
			roundRobin(id, conn, tries)
		case "active-polling", "passive-polling":
			polling(id, conn, tries)
		}
	} else {
		if config.Verbose {
			fmt.Printf("WARN - %s - Max retry cycles reached, aborting\n", id)
		}
		conn.Close()
	}
}

func forward(id string, conn net.Conn, upstr upstream) error {
	client, err := net.DialTimeout("tcp", upstr.IP+":"+upstr.Port, time.Duration(config.Timeout)*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN - %s - Forwarding to %s (%s:%s) failed: %v\n", id, upstr.Name, upstr.IP, upstr.Port, err)
		return errors.New("")
	}

	if config.Verbose {
		fmt.Printf("INFO - %s - Forwarding to %s (%s:%s)\n", id, upstr.Name, upstr.IP, upstr.Port)
	}
	go func() {
		defer client.Close()
		defer conn.Close()
		io.Copy(client, conn)
	}()
	go func() {
		defer client.Close()
		defer conn.Close()
		io.Copy(conn, client)
	}()

	if config.Sticky {
		remoteIP := stripPortFromAddr(conn.RemoteAddr().String())
		_, isCached := getSession(remoteIP)
		updateSession(remoteIP, sessionDetails{upstr, time.Now()})
		if config.Verbose {
			keyword := "Sticked"
			if isCached {
				keyword = "Resticked"
			}
			fmt.Printf("INFO - %s - %s %s to %s\n", id, keyword, remoteIP, upstr.Name)
		}
	}

	return nil
}

func stripPortFromAddr(remoteAddr string) string {
	i := strings.LastIndex(remoteAddr, ":")
	if i != -1 {
		remoteAddr = remoteAddr[:i]
	}
	return remoteAddr
}
