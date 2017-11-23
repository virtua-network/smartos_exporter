// vmstat collector
// this will :
//  - call vmstat
//  - gather memory metrics
//  - feed the collector

package collector

import (
	"os/exec"
	"strconv"
	"strings"
	// Prometheus Go toolset
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// GZFreeMemCollector declares the data type within the prometheus metrics package.
type GZFreeMemCollector struct {
	gzFreeMem *prometheus.GaugeVec
}

// NewGZFreeMemExporter returns a newly allocated exporter GZFreeMemCollector.
// It exposes the total free memory of the CN.
func NewGZFreeMemExporter() (*GZFreeMemCollector, error) {
	return &GZFreeMemCollector{
		gzFreeMem: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_memory_free_bytes",
			Help: "Total free memory (both RAM and Swap) of the CN.",
		}, []string{"memory"}),
	}, nil
}

// Describe describes all the metrics.
func (e *GZFreeMemCollector) Describe(ch chan<- *prometheus.Desc) {
	e.gzFreeMem.Describe(ch)
}

// Collect fetches the stats.
func (e *GZFreeMemCollector) Collect(ch chan<- prometheus.Metric) {
	e.vmstat()
	e.gzFreeMem.Collect(ch)
}

func (e *GZFreeMemCollector) vmstat() {
	// XXX needs enhancement :
	// use of vmstat will wait 2 seconds in order to collect statistics
	out, eerr := exec.Command("bash", "-c", "vmstat 1 2").Output()
	if eerr != nil {
		log.Errorf("error on executing vmstat: %v", eerr)
	}
	perr := e.parseVmstatOutput(string(out))
	if perr != nil {
		log.Errorf("error on parsing vmstat: %v", perr)
	}
}

func (e *GZFreeMemCollector) parseVmstatOutput(out string) error {
	outlines := strings.Split(out, "\n")
	l := len(outlines)
	for _, line := range outlines[3 : l-1] {
		parsedLine := strings.Fields(line)
		freeSwap, err := strconv.ParseFloat(parsedLine[3], 64)
		if err != nil {
			return err
		}
		freeRAM, err := strconv.ParseFloat(parsedLine[4], 64)
		if err != nil {
			return err
		}
		e.gzFreeMem.With(prometheus.Labels{"memory": "swap"}).Set(freeSwap)
		e.gzFreeMem.With(prometheus.Labels{"memory": "ram"}).Set(freeRAM)
	}
	return nil
}
