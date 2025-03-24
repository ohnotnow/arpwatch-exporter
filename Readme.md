# Arpwatch Exporter

A Prometheus exporter that parses ARP data from an `arpwatch`-generated file and exposes metrics on observed MAC/IP/hostname combinations.

## Features

- Exposes the following Prometheus metrics:
  - `arpwatch_device_last_seen_timestamp` (labels: `mac`, `ip`, `hostname`)
  - `arpwatch_exporter_read_errors_total`
  - `arpwatch_exporter_last_read_timestamp`
  - `arpwatch_devices_tracked_total`
- Periodically reads the `arp.dat` file every 30 seconds.
- Simple HTTP server exposing `/metrics` endpoint.

## Requirements

- Go 1.24+
- Prometheus client Golang library

## Installation

### Clone Repository

```bash
git clone https://github.com/yourusername/arpwatch-exporter.git
cd arpwatch-exporter
```

### Install Dependencies

```bash
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
```

## Compilation

To build the exporter for your current platform:

```bash
go build -o arpwatch_exporter arpwatch_exporter.go
```

### Cross-Compilation

You can compile for other platforms using the `GOOS` and `GOARCH` environment variables. For example:

```bash
# Linux x86_64
GOOS=linux GOARCH=amd64 go build -o arpwatch_exporter arpwatch_exporter.go

# Windows x86_64
GOOS=windows GOARCH=amd64 go build -o arpwatch_exporter.exe main.go

# ARM64 (e.g., Raspberry Pi)
GOOS=linux GOARCH=arm64 go build -o arpwatch_exporter_arm64 main.go
```

## Usage

```bash
./arpwatch_exporter \
  -web.listen-address=":9617" \
  -web.telemetry-path="/metrics" \
  -arpwatch.file="/var/lib/arpwatch/arp.dat"
```

### Flags

- `-web.listen-address` (default `:9617`) - Address to bind the HTTP server.
- `-web.telemetry-path` (default `/metrics`) - Path to expose Prometheus metrics.
- `-arpwatch.file` (default `/var/lib/arpwatch/arp.dat`) - Path to the ARP data file.

## Systemd Service (optional)

You may run the exporter as a systemd service:

```ini
[Unit]
Description=Arpwatch Prometheus Exporter
After=network.target

[Service]
ExecStart=/path/to/arpwatch_exporter -arpwatch.file=/var/lib/arpwatch/arp.dat
Restart=always

[Install]
WantedBy=multi-user.target
```

## License

This project is licensed under the MIT License.


