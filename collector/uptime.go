// uptime collector
// this will :
//  - call uptime
//  - gather load average
//  - feed the collector

package collector

import (
	"os/exec"
	"regexp"
	"strconv"
	// Prometheus Go toolset
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// LoadAverageCollector declares the data type within the prometheus metrics
// package.
type LoadAverageCollector struct {
	LoadAverage1  prometheus.Gauge
	LoadAverage5  prometheus.Gauge
	LoadAverage15 prometheus.Gauge
}

// NewLoadAverageExporter returns a newly allocated exporter LoadAverageCollector.
// It exposes the CPU load average.
func NewLoadAverageExporter() (*LoadAverageCollector, error) {
	return &LoadAverageCollector{
		LoadAverage1: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "smartos_cpu_load1",
			Help: "CPU load average 1 minute.",
		}),
		LoadAverage5: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "smartos_cpu_load5",
			Help: "CPU load average 5 minutes.",
		}),
		LoadAverage15: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "smartos_cpu_load15",
			Help: "CPU load average 15 minutes.",
		}),
	}, nil
}

// Describe describes all the metrics.
func (e *LoadAverageCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.LoadAverage1.Desc()
	ch <- e.LoadAverage5.Desc()
	ch <- e.LoadAverage15.Desc()
}

// Collect fetches the stats.
func (e *LoadAverageCollector) Collect(ch chan<- prometheus.Metric) {
	e.uptime()
	ch <- e.LoadAverage1
	ch <- e.LoadAverage5
	ch <- e.LoadAverage15
}

func (e *LoadAverageCollector) uptime() {
	out, eerr := exec.Command("bash", "-c", "uptime").Output()
	if eerr != nil {
		log.Errorf("error on executing uptime: %v", eerr)
	}
	perr := e.parseUptimeOutput(string(out))
	if perr != nil {
		log.Errorf("error on parsing uptime: %v", perr)
	}
}

func (e *LoadAverageCollector) parseUptimeOutput(out string) error {
	// we will use regex in order to be sure to catch good numbers
	r, _ := regexp.Compile(`load average: (\d+.\d+), (\d+.\d+), (\d+.\d+)`)
	loads := r.FindStringSubmatch(out)

	load1, err := strconv.ParseFloat(loads[1], 64)
	if err != nil {
		return err
	}
	load5, err := strconv.ParseFloat(loads[2], 64)
	if err != nil {
		return err
	}
	load15, err := strconv.ParseFloat(loads[3], 64)
	if err != nil {
		return err
	}

	e.LoadAverage1.Set(load1)
	e.LoadAverage5.Set(load5)
	e.LoadAverage15.Set(load15)

	return nil
}
