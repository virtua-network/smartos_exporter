// Virtua SmartOS Prometheus exporter
//
// Worflow :
//  - detect if launched in GZ or into a Zone
//  - retrieve useful metrics
//  - expose them to http://xxx:9100/metrics (same as node_exporter)

package main

import (
    "log"
    "net/http"
    "os/exec"
    "strconv"
    "strings"
//  "fmt"

    "github.com/virtua-network/smartos_exporter/collector"

    // Prometheus Go toolset
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    // Global variables
    exporterPort = ":9100"
    // Metrics definitions
    // global zone metrics
    gzZpoolList = prometheus.NewGaugeVec(prometheus.GaugeOpts{
        Name: "smartos_gz_zpool_list_total",
        Help: "ZFS zpool list summary.",
        },
        []string{"zpool","type"},
    )
    // zone metrics
)

// Global Helpers

// try to determine if its executed inside the GZ or not.
// return 1 if in GZ
//        0 if in zone
func isGlobalZone() (int) {
    out, eerr := exec.Command("zonename").Output()
    if eerr != nil {
        log.Fatal(eerr)
    }
    if (strings.Contains(string(out), "global")) == false {
        return 0
    } else {
        return 1
    }
}

// SmartOS command / tool callers

func zpoolList() {
    out, eerr := exec.Command("zpool", "list", "-p", "zones").Output()
    if eerr != nil {
        log.Fatal(eerr)
    }
    perr := parseZpoolStatusOutput(string(out))
    if perr != nil {
        log.Fatal(perr)
    }
}

// Parsers

func parseZpoolStatusOutput(out string) (error) {
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
          gzZpoolList.With(prometheus.Labels{"zpool":"zones", "type":"faulty"}).Set(0)
        } else {
          gzZpoolList.With(prometheus.Labels{"zpool":"zones", "type":"faulty"}).Set(1)
        }

        gzZpoolList.With(prometheus.Labels{"zpool":"zones", "type":"size"}).Set(sizeBytes)
        gzZpoolList.With(prometheus.Labels{"zpool":"zones", "type":"alloc"}).Set(allocBytes)
        gzZpoolList.With(prometheus.Labels{"zpool":"zones", "type":"free"}).Set(freeBytes)
        gzZpoolList.With(prometheus.Labels{"zpool":"zones", "type":"frag"}).Set(fragPercentTrim)
        gzZpoolList.With(prometheus.Labels{"zpool":"zones", "type":"capacity"}).Set(capPercentTrim)
    }
    return nil
}

// program starter

func init() {
    // Metrics have to be registered to be exposed:
    gz := isGlobalZone()
    if gz == 0 {
        // not yet implemented
        // XXX
        log.Fatal("zone statistics gathering is not yet implemented.")
    }
}

func main() {
    prometheus.MustRegister(gzZpoolList)

    freemem, _ := collector.NewGzFreeMemExporter()
    prometheus.MustRegister(freemem)

    mlagusage, _ := collector.NewGzMlagUsageExporter()
    prometheus.MustRegister(mlagusage)

    loadavg, _ := collector.NewLoadAverageExporter()
    prometheus.MustRegister(loadavg)

    cpuusage, _ := collector.NewGzCpuUsageExporter()
    prometheus.MustRegister(cpuusage)

    diskerrors, _ := collector.NewGzDiskErrorsExporter()
    prometheus.MustRegister(diskerrors)

    zpoolList()

    // The Handler function provides a default handler to expose metrics
    // via an HTTP server. "/metrics" is the usual endpoint for that.
    http.Handle("/metrics", promhttp.Handler())
    log.Fatal(http.ListenAndServe(exporterPort, nil))
}
