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
//      "fmt"
        // Prometheus Go toolset
        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    // Global variables
    exporterPort := ":9100"
    // Metrics definitions
    gzCpuUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
        Name: "smartos_gz_cpu_usage_total",
        Help: "CPU usage exposed in percent.",
        },
        []string{"cpu","type"},
    )
    gzFreeMem = prometheus.NewGaugeVec(prometheus.GaugeOpts{
        Name: "smartos_gz_memory_free_bytes_total",
        Help: "Total free memory (both RAM and Swap) of the CN.",
        },
        []string{"type"},
    )
    gzMlagUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
        Name: "smartos_gz_network_mlag_bytes_total",
        Help: "MLAG (aggr0) usage of the CN.",
        },
        []string{"device","type"},
    )
    gzZpoolList = prometheus.NewGaugeVec(prometheus.GaugeOpts{
        Name: "smartos_gz_zpool_list_total",
        Help: "ZFS zpool list summary.",
        },
        []string{"zpool","type"},
    )
    gzDiskErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
        Name: "smartos_gz_disk_errors_total",
        Help: "Number of hardware disk errors.",
        },
        []string{"device","type"},
    )
)

// program start

func init() {
    // Metrics have to be registered to be exposed:
    gz, _ := isGlobalZone()
    if gz == 1 {
        prometheus.MustRegister(gzCpuUsage)
        prometheus.MustRegister(gzFreeMem)
        prometheus.MustRegister(gzMlagUsage)
        prometheus.MustRegister(gzDiskErrors)
        prometheus.MustRegister(gzZpoolList)
    } else {
        // not yet implemented
        log.Fatal("not yet implemented")
    }
}

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
func iostat() {
    out, eerr := exec.Command("iostat", "-en").Output()
    if eerr != nil {
        log.Fatal(eerr)
    }
    perr := parseIostatOutput(string(out))
    if perr != nil {
        log.Fatal(perr)
    }
}

func mpstat() {
    out, eerr := exec.Command("mpstat", "1", "1").Output()
    if eerr != nil {
        log.Fatal(eerr)
    }
    perr := parseMpstatOutput(string(out))
    if perr != nil {
        log.Fatal(perr)
    }
}

func nicstat() {
    out, eerr := exec.Command("nicstat", "-i", "aggr0").Output()
    if eerr != nil {
        log.Fatal(eerr)
    }
    perr := parseNicstatOutput(string(out))
    if perr != nil {
        log.Fatal(perr)
    }
}

func vmstat() {
    out, eerr := exec.Command("vmstat", "1", "1").Output()
    if eerr != nil {
        log.Fatal(eerr)
    }
    perr := parseVmstatOutput(string(out))
    if perr != nil {
        log.Fatal(perr)
    }
}

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

func parseIostatOutput(out string) (error) {
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
        gzDiskErrors.With(prometheus.Labels{"device":deviceName,"type":"soft"}).Add(softErr)
        gzDiskErrors.With(prometheus.Labels{"device":deviceName,"type":"hard"}).Add(hardErr)
        gzDiskErrors.With(prometheus.Labels{"device":deviceName,"type":"trn"}).Add(trnErr)
    }
    return nil
}

func parseMpstatOutput(out string) (error) {
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
        gzCpuUsage.With(prometheus.Labels{"cpu": cpuId, "type":"user"}).Set(cpuUsr)
        gzCpuUsage.With(prometheus.Labels{"cpu": cpuId, "type":"system"}).Set(cpuSys)
        gzCpuUsage.With(prometheus.Labels{"cpu": cpuId, "type":"idle"}).Set(cpuIdl)
        //fmt.Printf("cpuId : %d, cpuUsr : %d, cpuSys : %d \n", cpuId, cpuUsr, cpuSys)
    }
    return nil
}

func parseNicstatOutput(out string) (error) {
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
        gzMlagUsage.With(prometheus.Labels{"device":"aggr0", "type":"read"}).Set(readKb)
        gzMlagUsage.With(prometheus.Labels{"device":"aggr0", "type":"write"}).Set(writeKb)
    }
    return nil
}

func parseVmstatOutput(out string) (error) {
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
        gzFreeMem.With(prometheus.Labels{"type":"swap"}).Set(freeSwap)
        gzFreeMem.With(prometheus.Labels{"type":"ram"}).Set(freeRam)
    }
    return nil
}

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
    gz, _ := isGlobalZone()
    if gz == 1 {
        prometheus.MustRegister(gzCpuUsage)
        prometheus.MustRegister(gzFreeMem)
        prometheus.MustRegister(gzMlagUsage)
        prometheus.MustRegister(gzDiskErrors)
        prometheus.MustRegister(gzZpoolList)
    } else {
        // not yet implemented
        // XXX
        log.Fatal("zone statistics gathering is not yet implemented.")
    }
}

func main() {
    ret, _ := isGlobalZone()
    if ret == 1 {
        mpstat()
        vmstat()
        nicstat()
        iostat()
        zpoolList()
    } else {
        // not yet implemented
        // XXX
        log.Fatal("zone statistics gathering is not yet implemented.")
    }

    // The Handler function provides a default handler to expose metrics
    // via an HTTP server. "/metrics" is the usual endpoint for that.
    http.Handle("/metrics", promhttp.Handler())
    log.Fatal(http.ListenAndServe(exporterPort, nil))
}
