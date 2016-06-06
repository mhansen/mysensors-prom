package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/buxtronix/mysensors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tarm/serial"
)

var (
	addr      = flag.String("listen", ":9001", "Address to listen on")
	baud      = flag.Int("baud", 115200, "Baud rate")
	port      = flag.String("port", "/dev/ttyUSB0", "Serial port to open")
	stateFile = flag.String("state_file", ".mysensors-state", "File to save/read state")
	configFile = flag.String("config_file", "mysensors.cfg", "File containing config")
)

var (
	gauge     *prometheus.GaugeVec
	humGauge     *prometheus.GaugeVec
	pressGauge     *prometheus.GaugeVec
	battGauge *prometheus.GaugeVec
)

var p *serial.Port

func main() {
	flag.Parse()
	http.Handle("/metrics", prometheus.Handler())

	gauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mysensors_temp",
			Help: "Mysensors temperature",
			ConstLabels:  prometheus.Labels{"instance": "192.168.0.10:9001"},
		},
		[]string{"location"},
	)
	prometheus.MustRegister(gauge)

	humGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mysensors_humidity",
			Help: "Mysensors humidity",
			ConstLabels:  prometheus.Labels{"instance": "192.168.0.10:9001"},
		},
		[]string{"location"},
	)
	prometheus.MustRegister(humGauge)

	pressGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mysensors_pressure",
			Help: "Mysensors pressure",
			ConstLabels:  prometheus.Labels{"instance": "192.168.0.10:9001"},
		},
		[]string{"location"},
	)
	prometheus.MustRegister(pressGauge)

	battGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mysensors_battery",
			Help: "Mysensors battery levels",
			ConstLabels:  prometheus.Labels{"instance": "192.168.0.10:9001"},
		},
		[]string{"location"},
	)
	prometheus.MustRegister(battGauge)

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
		log.Printf("Error loading state: %v", err)
	}

	h := mysensors.NewHandler(p, p, ch, net)

	go func() {
		if err := http.ListenAndServe(*addr, nil); err != nil {
			panic(err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		for _ = range sigCh {
//			if err = st.Save(*stateFile); err != nil {
//				log.Printf("Error writing state file [%s]: %v", *stateFile, err)
//			}
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
		if err := net.HandleMessage(m); err != nil {
			log.Printf("HandleMessage: %v\n", err)
		}
		switch m.Type {
		case mysensors.MsgSet:
			subType := m.SubType.(mysensors.SubTypeSetReq)
			switch subType {
			case mysensors.V_TEMP:
				v, err := strconv.ParseFloat(string(m.Payload), 64)
				if err != nil {
					log.Printf("Payload error: %v\n", err)
					continue
				}
				if m.NodeID == 1 {
					gauge.WithLabelValues("attic").Set(v)
				}
				if m.NodeID == 2 {
					gauge.WithLabelValues("office").Set(v)
				}
				if m.NodeID == 3 {
					gauge.WithLabelValues("roof").Set(v)
				}
				if m.NodeID == 4 {
					gauge.WithLabelValues("outside").Set(v)
				}
			case mysensors.V_HUM:
				v, err := strconv.ParseFloat(string(m.Payload), 64)
				if err != nil {
					log.Printf("Payload error: %v\n", err)
					continue
				}
				if m.NodeID == 4 {
					humGauge.WithLabelValues("outside").Set(v)
				}
			case mysensors.V_PRESSURE:
				v, err := strconv.ParseFloat(string(m.Payload), 64)
				if err != nil {
					log.Printf("Payload error: %v\n", err)
					continue
				}
				if m.NodeID == 4 {
					pressGauge.WithLabelValues("outside").Set(v)
				}
			}
			mqttCh <- m
		case mysensors.MsgInternal:
			subType := m.SubType.(mysensors.SubTypeInternal)
			switch subType {
			case mysensors.I_BATTERY_LEVEL:
				v, err := strconv.ParseFloat(string(m.Payload), 64)
				if err != nil {
					log.Printf("Payload error: %v\n", err)
					continue
				}
				if m.NodeID == 1 {
					battGauge.WithLabelValues("attic").Set(v)
				}
				if m.NodeID == 2 {
					battGauge.WithLabelValues("office").Set(v)
				}
				if m.NodeID == 3 {
					battGauge.WithLabelValues("roof").Set(v)
				}
				if m.NodeID == 4 {
					battGauge.WithLabelValues("outside").Set(v)
				}
				mqttCh <- m
			}
		}
	}
}
