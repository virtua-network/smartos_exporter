// mpstat collector
// this will :
//  - call mpstat
//  - gather CPU metrics
//  - feed the collector

package collector

import (
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	// Prometheus Go toolset
	"github.com/prometheus/client_golang/prometheus"
)

// GZCPUUsageCollector declare the data type within the prometheus metrics
// package.
type GZCPUUsageCollector struct {
	gzCPUUsage *prometheus.GaugeVec
}

// NewGZCPUUsageExporter returns a newly allocated exporter GZCPUUsageCollector.
// It exposes the CPU usage in percent.
func NewGZCPUUsageExporter() (*GZCPUUsageCollector, error) {
	return &GZCPUUsageCollector{
		gzCPUUsage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_cpu_usage_percents",
			Help: "CPU usage exposed in percent.",
		}, []string{"cpu", "mode"}),
	}, nil
}

// Describe describes all the metrics.
func (e *GZCPUUsageCollector) Describe(ch chan<- *prometheus.Desc) {
	e.gzCPUUsage.Describe(ch)
}

// Collect fetches the stats.
func (e *GZCPUUsageCollector) Collect(ch chan<- prometheus.Metric) {
	e.mpstat()
	e.gzCPUUsage.Collect(ch)
}

func (e *GZCPUUsageCollector) mpstat() {
	// XXX needs enhancement :
	// use of mpstat will wait 2 seconds in order to collect statistics
	out, eerr := exec.Command("mpstat", "1", "2").Output()
	if eerr != nil {
		fmt.Errorf("error on executing mpstat: %v", eerr)
	}
	perr := e.parseMpstatOutput(string(out))
	if perr != nil {
		fmt.Errorf("error on parsing mpstat: %v", perr)
	}
}

func (e *GZCPUUsageCollector) parseMpstatOutput(out string) error {
	// this regexp will remove all lines containing header labels
	r, _ := regexp.Compile(`(?m)[\r\n]+^.*CPU.*$`)
	result := r.ReplaceAllString(out, "")

	outlines := strings.Split(result, "\n")
	l := len(outlines)
	for _, line := range outlines[1 : l-1] {
		parsedLine := strings.Fields(line)
		cpuID := parsedLine[0]
		cpuUsr, err := strconv.ParseFloat(parsedLine[12], 64)
		if err != nil {
			return err
		}
		cpuSys, err := strconv.ParseFloat(parsedLine[13], 64)
		if err != nil {
			return err
		}
		cpuIdl, err := strconv.ParseFloat(parsedLine[15], 64)
		if err != nil {
			return err
		}
		e.gzCPUUsage.With(prometheus.Labels{"cpu": cpuID, "mode": "user"}).Set(cpuUsr)
		e.gzCPUUsage.With(prometheus.Labels{"cpu": cpuID, "mode": "system"}).Set(cpuSys)
		e.gzCPUUsage.With(prometheus.Labels{"cpu": cpuID, "mode": "idle"}).Set(cpuIdl)
		//fmt.Printf("cpuID : %d, cpuUsr : %d, cpuSys : %d \n", cpuID, cpuUsr, cpuSys)
	}
	return nil
}
