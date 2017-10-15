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

    zpoollist, _ := collector.NewGzZpoolListExporter()
    prometheus.MustRegister(zpoollist)

    // The Handler function provides a default handler to expose metrics
    // via an HTTP server. "/metrics" is the usual endpoint for that.
    http.Handle("/metrics", promhttp.Handler())
    log.Fatal(http.ListenAndServe(exporterPort, nil))
}
