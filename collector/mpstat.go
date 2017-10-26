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

type gzCPUUsageExporter struct {
	gzCPUUsage *prometheus.GaugeVec
}

func NewGZCPUUsageExporter() (*gzCPUUsageExporter, error) {
	return &gzCPUUsageExporter{
		gzCPUUsage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_cpu_usage_percents",
			Help: "CPU usage exposed in percent.",
		}, []string{"cpu", "mode"}),
	}, nil
}

func (e *gzCPUUsageExporter) Describe(ch chan<- *prometheus.Desc) {
	e.gzCPUUsage.Describe(ch)
}

func (e *gzCPUUsageExporter) Collect(ch chan<- prometheus.Metric) {
	e.mpstat()
	e.gzCPUUsage.Collect(ch)
}

func (e *gzCPUUsageExporter) mpstat() {
	// XXX needs enhancement :
	// use of mpstat will wait 2 seconds in order to collect statistics
	out, eerr := exec.Command("mpstat", "1", "2").Output()
	if eerr != nil {
		log.Fatal(eerr)
	}
	perr := e.parseMpstatOutput(string(out))
	if perr != nil {
		log.Fatal(perr)
	}
}

func (e *gzCPUUsageExporter) parseMpstatOutput(out string) error {
	// this regexp will remove all lines containing header labels
	r, _ := regexp.Compile(`(?m)[\r\n]+^.*CPU.*$`)
	result := r.ReplaceAllString(out, "")

	outlines := strings.Split(result, "\n")
	l := len(outlines)
	for _, line := range outlines[1 : l-1] {
		parsedLine := strings.Fields(line)
		cpuId := parsedLine[0]
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
		e.gzCPUUsage.With(prometheus.Labels{"cpu": cpuId, "mode": "user"}).Set(cpuUsr)
		e.gzCPUUsage.With(prometheus.Labels{"cpu": cpuId, "mode": "system"}).Set(cpuSys)
		e.gzCPUUsage.With(prometheus.Labels{"cpu": cpuId, "mode": "idle"}).Set(cpuIdl)
		//fmt.Printf("cpuId : %d, cpuUsr : %d, cpuSys : %d \n", cpuId, cpuUsr, cpuSys)
	}
	return nil
}
