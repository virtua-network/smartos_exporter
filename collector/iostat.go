// iostat collector
// this will :
//  - call iostat
//  - gather hard disk metrics
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

// GZDiskErrorsCollector declares the data type within the prometheus metrics package.
type GZDiskErrorsCollector struct {
	gzDiskErrors *prometheus.CounterVec
}

// NewGZDiskErrorsExporter returns a newly allocated exporter GZDiskErrorsCollector.
// It exposes the number of hardware disk errors
func NewGZDiskErrorsExporter() (*GZDiskErrorsCollector, error) {
	return &GZDiskErrorsCollector{
		gzDiskErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "smartos_disk_errs_total",
			Help: "Number of hardware disk errors.",
		}, []string{"device", "error_type"}),
	}, nil
}

// Describe describes all the metrics.
func (e *GZDiskErrorsCollector) Describe(ch chan<- *prometheus.Desc) {
	e.gzDiskErrors.Describe(ch)
}

// Collect fetches the stats.
func (e *GZDiskErrorsCollector) Collect(ch chan<- prometheus.Metric) {
	e.iostat()
	e.gzDiskErrors.Collect(ch)
}

func (e *GZDiskErrorsCollector) iostat() {
	out, eerr := exec.Command("bash", "-c", "iostat", "-en").Output()
	if eerr != nil {
		log.Errorf("error on executing iostat: %v", eerr)
	}
	perr := e.parseIostatOutput(string(out))
	if perr != nil {
		log.Errorf("error on parsing iostat: %v", perr)
	}
}

func (e *GZDiskErrorsCollector) parseIostatOutput(out string) error {
	outlines := strings.Split(out, "\n")
	l := len(outlines)
	for _, line := range outlines[2 : l-1] {
		parsedLine := strings.Fields(line)
		deviceName := parsedLine[4]
		softErr, err := strconv.ParseFloat(parsedLine[0], 64)
		if err != nil {
			return err
		}
		hardErr, err := strconv.ParseFloat(parsedLine[1], 64)
		if err != nil {
			return err
		}
		trnErr, err := strconv.ParseFloat(parsedLine[2], 64)
		if err != nil {
			return err
		}
		e.gzDiskErrors.With(prometheus.Labels{"device": deviceName, "error_type": "soft"}).Add(softErr)
		e.gzDiskErrors.With(prometheus.Labels{"device": deviceName, "error_type": "hard"}).Add(hardErr)
		e.gzDiskErrors.With(prometheus.Labels{"device": deviceName, "error_type": "trn"}).Add(trnErr)
	}
	return nil
}
