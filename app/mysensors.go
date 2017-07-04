package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/buxtronix/mysensors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tarm/serial"
)

var (
	addr       = flag.String("listen", ":9001", "Address to listen on")
	baud       = flag.Int("baud", 115200, "Baud rate")
	port       = flag.String("port", "/dev/ttyUSB0", "Serial port to open")
	stateFile  = flag.String("state_file", ".mysensors-state", "File to save/read state")
)

var p *serial.Port

func main() {
	flag.Parse()
	http.Handle("/metrics", prometheus.Handler())

	var err error

	c := &serial.Config{Name: *port, Baud: *baud}
	p, err = serial.OpenPort(c)
	if err != nil {
		log.Fatalf("Error opening serial port %s: %v", *port, err)
	}

	mqttCh := make(chan *mysensors.Message)
	mqtt := &mysensors.MQTTClient{}
	if err := mqtt.Start(mqttCh); err != nil {
		log.Fatalf("Error starting MQTT client: %v", err)
	}

	ch := make(chan *mysensors.Message)

	net := mysensors.NewNetwork()
	if err = net.LoadJson(*stateFile); err != nil {
		log.Fatalf("Error loading state: %v", err)
	}

	h := mysensors.NewHandler(p, p, ch, net)

	go func() {
		if err := http.ListenAndServe(*addr, nil); err != nil {
			panic(err)
		}
	}()

	// Catch SIGINT and save state before exiting.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		for _ = range sigCh {
			if err = net.SaveJson(*stateFile); err != nil {
				log.Printf("Error writing state file [%s]: %v", *stateFile, err)
			}
			os.Exit(0)
		}
	}()

	go h.Start()

	go func() {
		for {
			net.StatusString()
			time.Sleep(30 * time.Second)
		}
	}()

	for m := range ch {
		if err := net.HandleMessage(m, h.Tx); err != nil {
			log.Printf("HandleMessage: %v\n", err)
		}
	}
}
