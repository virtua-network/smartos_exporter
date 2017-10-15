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
    gzMLAGUsage      *prometheus.GaugeVec
}

func NewGZMLAGUsageExporter() (*gzMLAGUsageExporter, error) {
    return &gzMLAGUsageExporter{
        gzMLAGUsage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
            Name: "smartos_gz_network_mlag_bytes_total",
            Help: "MLAG (aggr0) usage of the CN.",
        }, []string{"device","type"}),
    }, nil
}

func (e *gzMLAGUsageExporter) Describe(ch chan<- *prometheus.Desc) {
    e.gzMLAGUsage.Describe(ch)
}

func (e *gzMLAGUsageExporter) Collect(ch chan<- prometheus.Metric) {
    e.nicstat()
    e.gzMLAGUsage.Collect(ch)
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

func (e *gzMLAGUsageExporter) parseNicstatOutput(out string) (error) {
    outlines := strings.Split(out, "\n")
    l := len(outlines)
    for _, line := range outlines[2:l-1] {
        parsedLine := strings.Fields(line)
        readKb, err := strconv.ParseFloat(parsedLine[2], 64)
        if err != nil {
            return err
        }
        writeKb, err := strconv.ParseFloat(parsedLine[3], 64)
        if err != nil {
            return err
        }
        e.gzMLAGUsage.With(prometheus.Labels{"device":"aggr0", "type":"read"}).Set(readKb)
        e.gzMLAGUsage.With(prometheus.Labels{"device":"aggr0", "type":"write"}).Set(writeKb)
    }
    return nil
}
