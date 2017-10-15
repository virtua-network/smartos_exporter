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

type gzMlagUsageExporter struct {
    gzMlagUsage      *prometheus.GaugeVec
}

func NewGzMlagUsageExporter() (*gzMlagUsageExporter, error) {
    return &gzMlagUsageExporter{
        gzMlagUsage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
            Name: "smartos_gz_network_mlag_bytes_total",
            Help: "MLAG (aggr0) usage of the CN.",
        }, []string{"device","type"}),
    }, nil
}

func (e *gzMlagUsageExporter) Describe(ch chan<- *prometheus.Desc) {
    e.gzMlagUsage.Describe(ch)
}

func (e *gzMlagUsageExporter) Collect(ch chan<- prometheus.Metric) {
    e.nicstat()
    e.gzMlagUsage.Collect(ch)
}

func (e *gzMlagUsageExporter) nicstat() {
    out, eerr := exec.Command("nicstat", "-i", "aggr0").Output()
    if eerr != nil {
        log.Fatal(eerr)
    }
    perr := e.parseNicstatOutput(string(out))
    if perr != nil {
        log.Fatal(perr)
    }
}

func (e *gzMlagUsageExporter) parseNicstatOutput(out string) (error) {
    outlines := strings.Split(out, "\n")
    l := len(outlines)
    for _, line := range outlines[1:l-1] {
        parsedLine := strings.Fields(line)
        readKb, err := strconv.ParseFloat(parsedLine[2], 64)
        if err != nil {
            return err
        }
        writeKb, err := strconv.ParseFloat(parsedLine[3], 64)
        if err != nil {
            return err
        }
        e.gzMlagUsage.With(prometheus.Labels{"device":"aggr0", "type":"read"}).Set(readKb)
        e.gzMlagUsage.With(prometheus.Labels{"device":"aggr0", "type":"write"}).Set(writeKb)
    }
    return nil
}
