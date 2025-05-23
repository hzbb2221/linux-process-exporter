package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/v3/process"
)

const (
	namespace = "process"
	version   = "1.0.0"
)

type WebConfig struct {
	TLSServerConfig struct {
		CertFile string `yaml:"cert_file"`
		KeyFile  string `yaml:"key_file"`
	} `yaml:"tls_server_config"`
	BasicAuthUsers map[string]string `yaml:"basic_auth_users"`
}

var (
	showVersion      = flag.Bool("version", false, "Show version information")
	showHelp         = flag.Bool("help", false, "Show help information")
	webConfigFile    = flag.String("web.config.file", "", "Path to configuration file for TLS and basic auth settings")
	webListenAddress = flag.String("web.listen-address", ":9113", "Address to listen on for web interface and telemetry")
)

type ProcessCollector struct {
	processPidDesc    *prometheus.Desc
	processCpuDesc    *prometheus.Desc
	processMemoryDesc *prometheus.Desc
}

func NewProcessCollector() *ProcessCollector {
	return &ProcessCollector{
		processPidDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "info"),
			"Process information with pid and name",
			[]string{"pid", "name"},
			nil,
		),
		processCpuDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "cpu_usage"),
			"Process CPU usage percentage",
			[]string{"pid", "name"},
			nil,
		),
		processMemoryDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "memory_usage"),
			"Process memory usage percentage",
			[]string{"pid", "name"},
			nil,
		),
	}
}

func (c *ProcessCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.processPidDesc
	ch <- c.processCpuDesc
	ch <- c.processMemoryDesc
}

func (c *ProcessCollector) Collect(ch chan<- prometheus.Metric) {
	processes, err := process.Processes()
	if err != nil {
		log.Printf("Error getting processes: %v", err)
		return
	}

	for _, p := range processes {
		pid := p.Pid
		name, err := p.Name()
		if err != nil {
			continue
		}

		// Process info metric
		ch <- prometheus.MustNewConstMetric(
			c.processPidDesc,
			prometheus.GaugeValue,
			float64(pid),
			strconv.Itoa(int(pid)),
			name,
		)

		// CPU usage metric
		cpuPercent, err := p.CPUPercent()
		if err == nil {
			ch <- prometheus.MustNewConstMetric(
				c.processCpuDesc,
				prometheus.GaugeValue,
				cpuPercent,
				strconv.Itoa(int(pid)),
				name,
			)
		}

		// Memory usage metric
		memPercent, err := p.MemoryPercent()
		if err == nil {
			ch <- prometheus.MustNewConstMetric(
				c.processMemoryDesc,
				prometheus.GaugeValue,
				float64(memPercent),
				strconv.Itoa(int(pid)),
				name,
			)
		}
	}
}

func basicAuth(handler http.Handler, users map[string]string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted")`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		expectedPass, exists := users[user]
		if !exists || expectedPass != pass {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("linux-process-exporter version %s\n", version)
		os.Exit(0)
	}

	if *showHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	collector := NewProcessCollector()
	prometheus.MustRegister(collector)

	handler := promhttp.Handler()

	if *webConfigFile != "" {
		data, err := ioutil.ReadFile(*webConfigFile)
		if err != nil {
			log.Fatalf("Error reading web config file: %v", err)
		}

		var config WebConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			log.Fatalf("Error parsing web config file: %v", err)
		}

		if len(config.BasicAuthUsers) > 0 {
			handler = basicAuth(handler, config.BasicAuthUsers)
			log.Printf("Basic authentication enabled")
		}

		if config.TLSServerConfig.CertFile != "" && config.TLSServerConfig.KeyFile != "" {
			log.Printf("Starting HTTPS server with TLS certificate on %s", *webListenAddress)
			http.Handle("/metrics", handler)
			log.Fatal(http.ListenAndServeTLS(*webListenAddress, config.TLSServerConfig.CertFile, config.TLSServerConfig.KeyFile, nil))
		}
	}

	log.Printf("Starting HTTP server on %s", *webListenAddress)
	http.Handle("/metrics", handler)
	log.Fatal(http.ListenAndServe(*webListenAddress, nil))
}
