// vmstat collector
// this will :
//  - call vmstat
//  - gather memory metrics
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

type gzFreeMemExporter struct {
    gzFreeMem      *prometheus.GaugeVec
}

func NewGzFreeMemExporter() (*gzFreeMemExporter, error) {
    return &gzFreeMemExporter{
        gzFreeMem: prometheus.NewGaugeVec(prometheus.GaugeOpts{
            Name: "smartos_gz_memory_free_bytes_total",
            Help: "Total free memory (both RAM and Swap) of the CN.",
        }, []string{"type"}),
    }, nil
}

func (e *gzFreeMemExporter) Describe(ch chan<- *prometheus.Desc) {
    e.gzFreeMem.Describe(ch)
}

func (e *gzFreeMemExporter) Collect(ch chan<- prometheus.Metric) {
    e.vmstat()
    e.gzFreeMem.Collect(ch)
}

func (e *gzFreeMemExporter) vmstat() {
    out, eerr := exec.Command("vmstat", "1", "1").Output()
    if eerr != nil {
        log.Fatal(eerr)
    }
    perr := e.parseVmstatOutput(string(out))
    if perr != nil {
        log.Fatal(perr)
    }
}

func (e *gzFreeMemExporter) parseVmstatOutput(out string) (error) {
    outlines := strings.Split(out, "\n")
    l := len(outlines)
    for _, line := range outlines[2:l-1] {
        parsedLine := strings.Fields(line)
        freeSwap, err := strconv.ParseFloat(parsedLine[3], 64)
        if err != nil {
            return err
        }
        freeRam, err := strconv.ParseFloat(parsedLine[4], 64)
        if err != nil {
            return err
        }
        e.gzFreeMem.With(prometheus.Labels{"type":"swap"}).Set(freeSwap)
        e.gzFreeMem.With(prometheus.Labels{"type":"ram"}).Set(freeRam)
    }
    return nil
}
