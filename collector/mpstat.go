// mpstat collector
// this will :
//  - call mpstat
//  - gather CPU metrics
//  - feed the collector

// XXX COLLECTOR BROKEN
// $(mpstat 1 1) always returns the same value

package collector

import (
    "log"
    "os/exec"
    "strconv"
    "strings"
    // Prometheus Go toolset
    "github.com/prometheus/client_golang/prometheus"
)

type gzCpuUsageExporter struct {
    gzCpuUsage      *prometheus.GaugeVec
}

func NewGzCpuUsageExporter() (*gzCpuUsageExporter, error) {
    return &gzCpuUsageExporter{
        gzCpuUsage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
            Name: "smartos_gz_cpu_usage_total",
            Help: "CPU usage exposed in percent.",
        }, []string{"cpu","type"}),
    }, nil
}

func (e *gzCpuUsageExporter) Describe(ch chan<- *prometheus.Desc) {
    e.gzCpuUsage.Describe(ch)
}

func (e *gzCpuUsageExporter) Collect(ch chan<- prometheus.Metric) {
    e.mpstat()
    e.gzCpuUsage.Collect(ch)
}

func (e *gzCpuUsageExporter) mpstat() {
    out, eerr := exec.Command("mpstat", "1", "1").Output()
    if eerr != nil {
        log.Fatal(eerr)
    }
    perr := e.parseMpstatOutput(string(out))
    if perr != nil {
        log.Fatal(perr)
    }
}

func (e *gzCpuUsageExporter) parseMpstatOutput(out string) (error) {
    outlines := strings.Split(out, "\n")
    l := len(outlines)
    for _, line := range outlines[1:l-1] {
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
        e.gzCpuUsage.With(prometheus.Labels{"cpu": cpuId, "type":"user"}).Set(cpuUsr)
        e.gzCpuUsage.With(prometheus.Labels{"cpu": cpuId, "type":"system"}).Set(cpuSys)
        e.gzCpuUsage.With(prometheus.Labels{"cpu": cpuId, "type":"idle"}).Set(cpuIdl)
        //fmt.Printf("cpuId : %d, cpuUsr : %d, cpuSys : %d \n", cpuId, cpuUsr, cpuSys)
    }
    return nil
}
