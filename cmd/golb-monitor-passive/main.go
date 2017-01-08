package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	ip      = kingpin.Flag("IP", "The address to listen on.").Short('i').Default("0.0.0.0").String()
	port    = kingpin.Flag("port", "The port to listen on.").Short('p').Default("1337").Int()
	cpuLoad []float64
)

func stats(res http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(res, "%f\n", cpuLoad[0])
}

func main() {
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()
	portString := strconv.Itoa(*port)

	go func() {
		for {
			var err error
			cpuLoad, err = cpu.Percent(0, false)
			if err != nil {
				cpuLoad = []float64{-1}
				kingpin.Fatalf("%v", err)
			}
			time.Sleep(3 * time.Second)
		}
	}()

	http.HandleFunc("/", stats)

	fmt.Printf("Listening on %s:%s...\n", *ip, portString)
	err := http.ListenAndServe(*ip+":"+portString, nil)
	kingpin.FatalIfError(err, "")
}
