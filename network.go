package golb

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
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

	logrus.WithFields(logrus.Fields{
		"host": host,
		"port": config.Port,
	}).Info("Listening...")

	for {
		conn, err := listener.Accept()
		id := uuid.NewV1().String()
		if err != nil || conn == nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
			}).Warn("Failed to accept listener")
			continue
		}

		if config.Verbose {
			logrus.WithFields(logrus.Fields{
				"id":             id,
				"remote-address": conn.RemoteAddr().String(),
			}).Info("Accepted connection")
		}

		go func() {
			if config.Sticky {
				if session, isCached := getSession(stripPortFromAddr(conn.RemoteAddr().String())); isCached {
					if config.Verbose {
						logrus.WithFields(logrus.Fields{
							"id":            id,
							"upstream-name": session.Upstream.Name,
						}).Info("Client is sticked to upstream")
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
			logrus.WithFields(logrus.Fields{
				"id": id,
			}).Warn("Max retry cycles reached, aborting")
		}
		drop(conn)
	}
}

func forward(id string, conn net.Conn, upstr upstream) error {
	client, err := net.DialTimeout("tcp", upstr.IP+":"+upstr.Port, time.Duration(config.Timeout)*time.Second)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id":            id,
			"err":           err,
			"upstream-name": upstr.Name,
		}).Warn("Forwarding failed")
		return errors.New("")
	}

	if config.Verbose {
		logrus.WithFields(logrus.Fields{
			"id":            id,
			"upstream-name": upstr.Name,
		}).Info("Forwarding succeeded")
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
			logrus.WithFields(logrus.Fields{
				"id":             id,
				"remote-address": remoteIP,
				"upstream-name":  upstr.Name,
			}).Infof("%s client to upstream", keyword)
		}
	}

	return nil
}

func drop(conn net.Conn) {
	conn.Close()
}

func stripPortFromAddr(remoteAddr string) string {
	i := strings.LastIndex(remoteAddr, ":")
	if i != -1 {
		remoteAddr = remoteAddr[:i]
	}
	return remoteAddr
}
