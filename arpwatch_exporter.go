package main

import (
	"bufio"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Command line flags
	listenAddress = flag.String("web.listen-address", ":9617", "Address to listen on for telemetry")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics")
	arpwatchFile  = flag.String("arpwatch.file", "/var/lib/arpwatch/arp.dat", "Path to the arpwatch data file")
	
	// Prometheus metrics
	lastSeenTimestamp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "arpwatch_device_last_seen_timestamp",
			Help: "Unix timestamp when a MAC address was last seen",
		},
		[]string{"mac", "ip", "hostname"},
	)
	
	fileReadErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "arpwatch_exporter_read_errors_total",
			Help: "Total number of arpwatch file read errors",
		},
	)
	
	lastFileReadTime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "arpwatch_exporter_last_read_timestamp",
			Help: "Unix timestamp of the last successful file read",
		},
	)
	
	devicesTracked = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "arpwatch_devices_tracked_total",
			Help: "Total number of devices currently being tracked",
		},
	)
)

func init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(lastSeenTimestamp)
	prometheus.MustRegister(fileReadErrors)
	prometheus.MustRegister(lastFileReadTime)
	prometheus.MustRegister(devicesTracked)
}

func readArpwatchData(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening arpwatch file: %v", err)
		fileReadErrors.Inc()
		return
	}
	defer file.Close()

	// Clear previous data before updating
	lastSeenTimestamp.Reset()
	
	scanner := bufio.NewScanner(file)
	deviceCount := 0
	
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}
		
		parts := strings.Fields(line)
		if len(parts) < 3 {
			log.Printf("Invalid line format: %s", line)
			continue
		}
		
		mac := parts[0]
		ip := parts[1]
		timestamp, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			log.Printf("Invalid timestamp format: %s", parts[2])
			continue
		}
		
		// Include hostname as a label if available
		hostname := ""
		if len(parts) >= 4 {
			hostname = parts[3]
		}
		lastSeenTimestamp.WithLabelValues(mac, ip, hostname).Set(float64(timestamp))
		deviceCount++
	}
	
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading arpwatch file: %v", err)
		fileReadErrors.Inc()
		return
	}
	
	devicesTracked.Set(float64(deviceCount))
	lastFileReadTime.Set(float64(time.Now().Unix()))
}

func updateMetrics() {
	for {
		readArpwatchData(*arpwatchFile)
		time.Sleep(30 * time.Second)
	}
}

func main() {
	flag.Parse()
	
	log.Printf("Starting arpwatch exporter on %s", *listenAddress)
	log.Printf("Metrics available at %s%s", *listenAddress, *metricsPath)
	log.Printf("Reading arpwatch data from %s", *arpwatchFile)
	
	go updateMetrics()
	
	// Expose the registered metrics via HTTP
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Arpwatch Exporter</title></head>
			<body>
			<h1>Arpwatch Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})
	
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

