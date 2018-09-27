# MySensors prometheus exporter.

A service to export MySensors data via Prometheus metrics.

## Mysensors?

MySensors (http://www.mysensors.org) is a project focussed on DIY home
automation and IoT.

## Prometheus?

Prometheus (http://prometheus.io) implements monitoring, alerting and
visualisation of data within a time series database.

## What's this package?

This package connects to a MySensors gateway over a serial port and
exports the received metrics for Prometheus to scrape.

```

  ^     +----------+    +------------------------+
  |     |  Serial  |    |    Host                |
  ------| Gateway  |----|--+------------------+  |
        +----------+    |  |  mysensors-prom  |--+--> prometheus
                        |  +------------------+  |
                        +------------------------+

```

As a serial gateway is relatively dumb, this package also handles
much of the internal MySensors logic, such as tracking and assigning
node IDs.

## Usage

Build a MySensors Serial Gateway.

Build the binary:

`go build app/mysensors.go`

Connect the Gateway.

Run the binary (see --help for other flags):

`./mysensors`

Metrics are then visible on http://localhost:9001/ as they
are received.

