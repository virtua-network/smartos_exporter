// df collector
// this will :
//  - call df
//  - gather disk free space metrics
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

// ZoneDfCollector declares the data type within the prometheus metrics package.
type ZoneDfCollector struct {
	ZoneDfSize      *prometheus.GaugeVec
	ZoneDfUsed      *prometheus.GaugeVec
	ZoneDfAvailable *prometheus.GaugeVec
	ZoneDfUse       *prometheus.GaugeVec
}

// NewZoneDfExporter returns a newly allocated exporter ZoneDfCollector.
// It exposes the df command result.
func NewZoneDfExporter() (*ZoneDfCollector, error) {
	return &ZoneDfCollector{
		ZoneDfSize: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_df_size_bytes",
			Help: "disk size in bytes.",
		}, []string{"device", "mountpoint"}),
		ZoneDfUsed: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_df_used_bytes",
			Help: "disk used space in bytes.",
		}, []string{"device", "mountpoint"}),
		ZoneDfAvailable: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_df_available_bytes",
			Help: "disk available space in bytes.",
		}, []string{"device", "mountpoint"}),
		ZoneDfUse: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_df_use_percents",
			Help: "disk used space in percents.",
		}, []string{"device", "mountpoint"}),
	}, nil
}

// Describe describes all the metrics.
func (e *ZoneDfCollector) Describe(ch chan<- *prometheus.Desc) {
	e.ZoneDfSize.Describe(ch)
	e.ZoneDfUsed.Describe(ch)
	e.ZoneDfAvailable.Describe(ch)
	e.ZoneDfUse.Describe(ch)
}

// Collect fetches the stats.
func (e *ZoneDfCollector) Collect(ch chan<- prometheus.Metric) {
	e.dfList()
	e.ZoneDfSize.Collect(ch)
	e.ZoneDfUsed.Collect(ch)
	e.ZoneDfAvailable.Collect(ch)
	e.ZoneDfUse.Collect(ch)
}

func (e *ZoneDfCollector) dfList() {
	out, eerr := exec.Command("df").Output()
	if eerr != nil {
		log.Errorf("error on executing df: %v", eerr)
	}
	perr := e.parseDfListOutput(string(out))
	if perr != nil {
		log.Errorf("error on parsing df output: %v", perr)
	}
}

func (e *ZoneDfCollector) parseDfListOutput(out string) error {
	outlines := strings.Split(out, "\n")
	l := len(outlines)
	// skip the first line (labels)
	for _, line := range outlines[1 : l-1] {
		parsedLine := strings.Fields(line)
		deviceName := parsedLine[0]
		mountName := parsedLine[5]
		sizeBytes, err := strconv.ParseFloat(parsedLine[1], 64)
		if err != nil {
			return err
		}
		usedBytes, err := strconv.ParseFloat(parsedLine[2], 64)
		if err != nil {
			return err
		}
		availBytes, err := strconv.ParseFloat(parsedLine[3], 64)
		if err != nil {
			return err
		}
		usePercent := strings.TrimSuffix(parsedLine[4], "%")
		usePercentTrim, err := strconv.ParseFloat(usePercent, 64)
		if err != nil {
			return err
		}

		e.ZoneDfSize.With(prometheus.Labels{"device": deviceName, "mountpoint": mountName}).Set(sizeBytes)
		e.ZoneDfUsed.With(prometheus.Labels{"device": deviceName, "mountpoint": mountName}).Set(usedBytes)
		e.ZoneDfAvailable.With(prometheus.Labels{"device": deviceName, "mountpoint": mountName}).Set(availBytes)
		e.ZoneDfUse.With(prometheus.Labels{"device": deviceName, "mountpoint": mountName}).Set(usePercentTrim)
	}
	return nil
}
