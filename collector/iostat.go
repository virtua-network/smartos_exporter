// iostat collector
// this will :
//  - call iostat
//  - gather hard disk metrics
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

type gzDiskErrorsExporter struct {
    gzDiskErrors      *prometheus.CounterVec
}

func NewGZDiskErrorsExporter() (*gzDiskErrorsExporter, error) {
    return &gzDiskErrorsExporter{
        gzDiskErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
            Name: "smartos_disk_errs_total",
            Help: "Number of hardware disk errors.",
        }, []string{"device","error_type"}),
    }, nil
}

func (e *gzDiskErrorsExporter) Describe(ch chan<- *prometheus.Desc) {
    e.gzDiskErrors.Describe(ch)
}

func (e *gzDiskErrorsExporter) Collect(ch chan<- prometheus.Metric) {
    e.iostat()
    e.gzDiskErrors.Collect(ch)
}

func (e *gzDiskErrorsExporter) iostat() {
    out, eerr := exec.Command("iostat", "-en").Output()
    if eerr != nil {
        log.Fatal(eerr)
    }
    perr := e.parseIostatOutput(string(out))
    if perr != nil {
        log.Fatal(perr)
    }
}

func (e *gzDiskErrorsExporter) parseIostatOutput(out string) (error) {
    outlines := strings.Split(out, "\n")
    l := len(outlines)
    for _, line := range outlines[2:l-1] {
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
        e.gzDiskErrors.With(prometheus.Labels{"device":deviceName,"type":"soft"}).Add(softErr)
        e.gzDiskErrors.With(prometheus.Labels{"device":deviceName,"type":"hard"}).Add(hardErr)
        e.gzDiskErrors.With(prometheus.Labels{"device":deviceName,"type":"trn"}).Add(trnErr)
    }
    return nil
}
