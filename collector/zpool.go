// zpool collector
// this will :
//  - call zpool list 
//  - gather ZPOOL metrics
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

type gzZpoolListExporter struct {
    gzZpoolList      *prometheus.GaugeVec
}

func NewGzZpoolListExporter() (*gzZpoolListExporter, error) {
    return &gzZpoolListExporter{
        gzZpoolList: prometheus.NewGaugeVec(prometheus.GaugeOpts{
            Name: "smartos_gz_zpool_list_total",
            Help: "ZFS zpool list summary.",
        }, []string{"zpool","type"}),
    }, nil
}

func (e *gzZpoolListExporter) Describe(ch chan<- *prometheus.Desc) {
    e.gzZpoolList.Describe(ch)
}

func (e *gzZpoolListExporter) Collect(ch chan<- prometheus.Metric) {
    e.zpoolList()
    e.gzZpoolList.Collect(ch)
}

func (e *gzZpoolListExporter) zpoolList() {
    out, eerr := exec.Command("zpool", "list", "-p", "zones").Output()
    if eerr != nil {
        log.Fatal(eerr)
    }
    perr := e.parseZpoolListOutput(string(out))
    if perr != nil {
        log.Fatal(perr)
    }
}

func (e *gzZpoolListExporter) parseZpoolListOutput(out string) (error) {
    outlines := strings.Split(out, "\n")
    l := len(outlines)
    for _, line := range outlines[1:l-1] {
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
          e.gzZpoolList.With(prometheus.Labels{"zpool":"zones", "type":"faulty"}).Set(0)
        } else {
          e.gzZpoolList.With(prometheus.Labels{"zpool":"zones", "type":"faulty"}).Set(1)
        }

        e.gzZpoolList.With(prometheus.Labels{"zpool":"zones", "type":"size"}).Set(sizeBytes)
        e.gzZpoolList.With(prometheus.Labels{"zpool":"zones", "type":"alloc"}).Set(allocBytes)
        e.gzZpoolList.With(prometheus.Labels{"zpool":"zones", "type":"free"}).Set(freeBytes)
        e.gzZpoolList.With(prometheus.Labels{"zpool":"zones", "type":"frag"}).Set(fragPercentTrim)
        e.gzZpoolList.With(prometheus.Labels{"zpool":"zones", "type":"capacity"}).Set(capPercentTrim)
    }
    return nil
}
