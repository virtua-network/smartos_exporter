// mpstat collector
// this will :
//  - call mpstat
//  - gather CPU metrics
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

type gzCpuUsageExporter struct {
    gzCpuUsage      *prometheus.GaugeVec
}

//type LoadAverageExporter struct {
//    LoadAverage1    prometheus.Gauge
//    LoadAverage5    prometheus.Gauge
//    LoadAverage15   prometheus.Gauge
//}

func NewGzCpuUsageExporter() (*gzCpuUsageExporter, error) {
    return &gzCpuUsageExporter{
        gzCpuUsage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
            Name: "smartos_gz_cpu_usage_total",
            Help: "CPU usage exposed in percent.",
        }, []string{"cpu","type"}),
    }, nil
}

//func NewLoadAverageExporter() (*LoadAverageExporter, error) {
//    return &LoadAverageExporter{
//        LoadAverage1: prometheus.NewGauge(prometheus.GaugeOpts{
//            Name: "smartos_cpu_load1",
//            Help: "CPU load average 1 minute.",
//        }),
//        LoadAverage5: prometheus.NewGauge(prometheus.GaugeOpts{
//            Name: "smartos_cpu_load5",
//            Help: "CPU load average 5 minutes.",
//        }),
//        LoadAverage15: prometheus.NewGauge(prometheus.GaugeOpts{
//            Name: "smartos_cpu_load15",
//            Help: "CPU load average 15 minutes.",
//        }),
//    }, nil
//}

func (e *gzCpuUsageExporter) Describe(ch chan<- *prometheus.Desc) {
    e.gzCpuUsage.Describe(ch)
}

//func (e *LoadAverageExporter) Describe(ch chan<- *prometheus.Desc) {
//    ch <- e.LoadAverage1.Desc()
//    ch <- e.LoadAverage5.Desc()
//    ch <- e.LoadAverage15.Desc()
//}

func (e *gzCpuUsageExporter) Collect(ch chan<- prometheus.Metric) {
    e.mpstat()
    e.gzCpuUsage.Collect(ch)
}
//func (e *LoadAverageExporter) Collect(ch chan<- prometheus.Metric) {
//    e.uptime()
//    ch <- e.LoadAverage1
//    ch <- e.LoadAverage5
//    ch <- e.LoadAverage15
//}

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

//func (e *LoadAverageExporter) uptime() {
//    out, eerr := exec.Command("uptime").Output()
//    if eerr != nil {
//        log.Fatal(eerr)
//    }
//    perr := e.parseUptimeOutput(string(out))
//    if perr != nil {
//        log.Fatal(perr)
//    }
//}

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

//func (e *LoadAverageExporter) parseUptimeOutput(out string) (error) {
//    // we will use regex in order to be sure to catch good numbers
//    r,_ := regexp.Compile(`load average: (\d.\d+), (\d.\d+), (\d.\d+)`)
//    loads := r.FindStringSubmatch(out)
//
//    load1, err := strconv.ParseFloat(loads[1], 64)
//    if err != nil {
//        return err
//    }
//    load5, err := strconv.ParseFloat(loads[2], 64)
//    if err != nil {
//        return err
//    }
//    load15, err := strconv.ParseFloat(loads[3], 64)
//    if err != nil {
//        return err
//    }
//
//    e.LoadAverage1.Set(load1)
//    e.LoadAverage5.Set(load5)
//    e.LoadAverage15.Set(load15)
//
//    return nil
//}
