// nicstat collector
// this will :
//  - call nicstat
//  - gather network metrics
//  - feed the collector

package collector

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
	// Prometheus Go toolset
	"github.com/prometheus/client_golang/prometheus"
)

type gzMLAGUsageExporter struct {
	gzMLAGUsageRead  *prometheus.GaugeVec
	gzMLAGUsageWrite *prometheus.GaugeVec
}

func NewGZMLAGUsageExporter() (*gzMLAGUsageExporter, error) {
	return &gzMLAGUsageExporter{
		gzMLAGUsageRead: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_network_mlag_receive_kilobytes",
			Help: "MLAG (aggr0) receive statistic in KBytes.",
		}, []string{"device"}),
		gzMLAGUsageWrite: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_network_mlag_transmit_kilobytes",
			Help: "MLAG (aggr0) transmit statistic in KBytes.",
		}, []string{"device"}),
	}, nil
}

func (e *gzMLAGUsageExporter) Describe(ch chan<- *prometheus.Desc) {
	e.gzMLAGUsageRead.Describe(ch)
	e.gzMLAGUsageWrite.Describe(ch)
}

func (e *gzMLAGUsageExporter) Collect(ch chan<- prometheus.Metric) {
	e.nicstat()
	e.gzMLAGUsageRead.Collect(ch)
	e.gzMLAGUsageWrite.Collect(ch)
}

func (e *gzMLAGUsageExporter) nicstat() {
	// XXX needs enhancement :
	// use of nicstat will wait 2 seconds in order to collect statistics
	out, eerr := exec.Command("nicstat", "-i", "aggr0", "1", "2").Output()
	if eerr != nil {
		log.Fatal(eerr)
	}
	perr := e.parseNicstatOutput(string(out))
	if perr != nil {
		log.Fatal(perr)
	}
}

func (e *gzMLAGUsageExporter) parseNicstatOutput(out string) error {
	outlines := strings.Split(out, "\n")
	l := len(outlines)
	for _, line := range outlines[2 : l-1] {
		parsedLine := strings.Fields(line)
		readKb, err := strconv.ParseFloat(parsedLine[2], 64)
		if err != nil {
			return err
		}
		writeKb, err := strconv.ParseFloat(parsedLine[3], 64)
		if err != nil {
			return err
		}
		e.gzMLAGUsageRead.With(prometheus.Labels{"device": "aggr0"}).Set(readKb)
		e.gzMLAGUsageWrite.With(prometheus.Labels{"device": "aggr0"}).Set(writeKb)
	}
	return nil
}
