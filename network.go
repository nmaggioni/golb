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
	listener, err := net.Listen("tcp", config.IP+":"+config.Port)
	if err != nil {
		return fmt.Errorf("Failed to setup listener: %v", err)
	}

	host := config.IP
	if host == "" {
		host = "127.0.0.1"
	}
	fmt.Printf("Listening on %s:%s...\n", host, config.Port)

	for {
		conn, err := listener.Accept()
		id := uuid.NewV1().String()
		if err != nil {
			fmt.Errorf("Failed to accept listener: %v", err)
		}

		if config.Verbose {
			fmt.Printf("%s - Accepted connection from %v\n", id, conn.RemoteAddr())
		}
		go forwardWithStrategy(id, conn)
	}
}

func forwardWithStrategy(id string, conn net.Conn) {
	switch config.Strategy {
	case "round-robin":
		roundRobin(id, conn, 0)
	}
}

func forward(id string, conn net.Conn, upstr upstream) error {
	client, err := net.DialTimeout("tcp", upstr.IP+":"+upstr.Port, time.Duration(config.Timeout)*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s - Forwarding to %s (%s:%s) failed: %v\n", id, upstr.Name, upstr.IP, upstr.Port, err)
		return errors.New("")
	}

	if config.Verbose {
		fmt.Printf("%s - Forwarding to %s (%s:%s)\n", id, upstr.Name, upstr.IP, upstr.Port)
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

	return nil
}
