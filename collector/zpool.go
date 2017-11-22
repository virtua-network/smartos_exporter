// zpool collector
// this will :
//  - call zpool list
//  - gather ZPOOL metrics
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

// GZZpoolListCollector declares the data type within the prometheus metrics package.
type GZZpoolListCollector struct {
	gzZpoolListAlloc    *prometheus.GaugeVec
	gzZpoolListCapacity *prometheus.GaugeVec
	gzZpoolListFaulty   *prometheus.GaugeVec
	gzZpoolListFrag     *prometheus.GaugeVec
	gzZpoolListFree     *prometheus.GaugeVec
	gzZpoolListSize     *prometheus.GaugeVec
}

// NewGZZpoolListExporter returns a newly allocated exporter GZZpoolListCollector.
// It exposes the zpool list command result.
func NewGZZpoolListExporter() (*GZZpoolListCollector, error) {
	return &GZZpoolListCollector{
		gzZpoolListAlloc: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_zpool_alloc_bytes",
			Help: "ZFS zpool allocated size in bytes.",
		}, []string{"zpool"}),
		gzZpoolListCapacity: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_zpool_cap_percents",
			Help: "ZFS zpool capacity in percents.",
		}, []string{"zpool"}),
		gzZpoolListFaulty: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_zpool_faults",
			Help: "ZFS zpool health status.",
		}, []string{"zpool"}),
		gzZpoolListFrag: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_zpool_frag_percents",
			Help: "ZFS zpool fragmentation in percents.",
		}, []string{"zpool"}),
		gzZpoolListFree: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_zpool_free_bytes",
			Help: "ZFS zpool space available in bytes.",
		}, []string{"zpool"}),
		gzZpoolListSize: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_zpool_size_bytes",
			Help: "ZFS zpool allocated size in bytes.",
		}, []string{"zpool"}),
	}, nil
}

// Describe describes all the metrics.
func (e *GZZpoolListCollector) Describe(ch chan<- *prometheus.Desc) {
	e.gzZpoolListAlloc.Describe(ch)
	e.gzZpoolListCapacity.Describe(ch)
	e.gzZpoolListFaulty.Describe(ch)
	e.gzZpoolListFrag.Describe(ch)
	e.gzZpoolListFree.Describe(ch)
	e.gzZpoolListSize.Describe(ch)
}

// Collect fetches the stats.
func (e *GZZpoolListCollector) Collect(ch chan<- prometheus.Metric) {
	e.zpoolList()
	e.gzZpoolListAlloc.Collect(ch)
	e.gzZpoolListCapacity.Collect(ch)
	e.gzZpoolListFaulty.Collect(ch)
	e.gzZpoolListFrag.Collect(ch)
	e.gzZpoolListFree.Collect(ch)
	e.gzZpoolListSize.Collect(ch)
}

func (e *GZZpoolListCollector) zpoolList() {
	out, eerr := exec.Command("zpool", "list", "-p", "zones").Output()
	if eerr != nil {
		log.Errorf("error on executing zpool: %v", eerr)
	}
	perr := e.parseZpoolListOutput(string(out))
	if perr != nil {
		log.Errorf("error on parsing zpool: %v", perr)
	}
}

func (e *GZZpoolListCollector) parseZpoolListOutput(out string) error {
	outlines := strings.Split(out, "\n")
	l := len(outlines)
	for _, line := range outlines[1 : l-1] {
		parsedLine := strings.Fields(line)
		sizeBytes, err := strconv.ParseFloat(parsedLine[1], 64)
		if err != nil {
			return err
		}
		allocBytes, err := strconv.ParseFloat(parsedLine[2], 64)
		if err != nil {
			return err
		}
		freeBytes, err := strconv.ParseFloat(parsedLine[3], 64)
		if err != nil {
			return err
		}
		fragPercent := strings.TrimSuffix(parsedLine[5], "%")
		fragPercentTrim, err := strconv.ParseFloat(fragPercent, 64)
		if err != nil {
			return err
		}
		capPercent := strings.TrimSuffix(parsedLine[6], "%")
		capPercentTrim, err := strconv.ParseFloat(capPercent, 64)
		if err != nil {
			return err
		}
		health := parsedLine[8]
		if (strings.Contains(health, "ONLINE")) == true {
			e.gzZpoolListFaulty.With(prometheus.Labels{"zpool": "zones"}).Set(0)
		} else {
			e.gzZpoolListFaulty.With(prometheus.Labels{"zpool": "zones"}).Set(1)
		}

		e.gzZpoolListAlloc.With(prometheus.Labels{"zpool": "zones"}).Set(allocBytes)
		e.gzZpoolListCapacity.With(prometheus.Labels{"zpool": "zones"}).Set(capPercentTrim)
		e.gzZpoolListFrag.With(prometheus.Labels{"zpool": "zones"}).Set(fragPercentTrim)
		e.gzZpoolListFree.With(prometheus.Labels{"zpool": "zones"}).Set(freeBytes)
		e.gzZpoolListSize.With(prometheus.Labels{"zpool": "zones"}).Set(sizeBytes)
	}
	return nil
}
