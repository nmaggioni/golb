package golb

import (
	"bytes"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

var polledLoads map[string]float64

func pollLoad(client *http.Client, upstr upstream) float64 {
	response, err := client.Get("http://" + upstr.IP + ":" + upstr.MonitoringPort + "/")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":           err,
			"upstream-name": upstr.Name,
		}).Warn("Failed to poll upstream")
		return -1
	}
	defer response.Body.Close()

	buf := bytes.NewBuffer(make([]byte, 0, response.ContentLength))
	_, err = buf.ReadFrom(response.Body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":           err,
			"upstream-name": upstr.Name,
		}).Warn("Failed to read upstream's poll")
		return -1
	}
	body := string(buf.Bytes())
	load, err := strconv.ParseFloat(strings.TrimSpace(body), 64)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":           err,
			"upstream-name": upstr.Name,
		}).Warn("Failed to parse upstream's poll")
		return -1
	}
	return load
}

func pollLoads() {
	var client = &http.Client{
		Timeout: time.Second * time.Duration(config.Timeout),
	}
	for _, upstr := range config.Upstreams {
		polledLoads[upstr.Name] = pollLoad(client, upstr)
	}
}

func polling(id string, conn net.Conn, tries int) {
	lessLoaded := -1
	for i, upstr := range config.Upstreams {
		if polledLoads[upstr.Name] != -1 {
			if lessLoaded == -1 {
				lessLoaded = i
			} else if polledLoads[upstr.Name] < polledLoads[config.Upstreams[lessLoaded].Name] {
				lessLoaded = i
			}
		}
	}
	if lessLoaded == -1 {
		logrus.WithFields(logrus.Fields{
			"id": id,
		}).Warn("No polls available, dropping connection")
		drop(conn)
		return
	}

	err := forward(id, conn, config.Upstreams[lessLoaded])
	if err != nil {
		forwardWithStrategy(id, conn, tries+1)
	}
}
