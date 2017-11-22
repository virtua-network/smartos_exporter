// kstat zone collector
// this will :
//  - call kstat inside a zone
//  - gather zone metrics
//  - feed the collector

package collector

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	// Prometheus Go toolset
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// ZoneKstatCollector declares the data type within the prometheus metrics package.
type ZoneKstatCollector struct {
	ZoneKstatCPUBaseline   *prometheus.GaugeVec
	ZoneKstatCPUCap        *prometheus.GaugeVec
	ZoneKstatCPUMaxUsage   *prometheus.GaugeVec
	ZoneKstatCPUUsage      *prometheus.GaugeVec
	ZoneKstatMemCap        *prometheus.GaugeVec
	ZoneKstatMemFree       *prometheus.GaugeVec
	ZoneKstatMemNover      *prometheus.GaugeVec
	ZoneKstatMemPagedOut   *prometheus.GaugeVec
	ZoneKstatMemRSS        *prometheus.GaugeVec
	ZoneKstatNICCollisions *prometheus.GaugeVec
	ZoneKstatNICIErrors    *prometheus.GaugeVec
	ZoneKstatNICIPackets   *prometheus.GaugeVec
	ZoneKstatNICLinkState  *prometheus.GaugeVec
	ZoneKstatNICOBytes     *prometheus.GaugeVec
	ZoneKstatNICOErrors    *prometheus.GaugeVec
	ZoneKstatNICOPackets   *prometheus.GaugeVec
	ZoneKstatNICRBytes     *prometheus.GaugeVec
	ZoneKstatSwapCap       *prometheus.GaugeVec
	ZoneKstatSwapFree      *prometheus.GaugeVec
	ZoneKstatSwapUsed      *prometheus.GaugeVec
}

// ZoneKstatNIC defines the mapping of kstat link structure.
type ZoneKstatNIC struct {
	ifName, ifLabel string
}

// ZoneKstatNICs slice for key iteration
type ZoneKstatNICs []ZoneKstatNIC

// NewZoneKstatExporter returns a newly allocated exporter ZoneKstatCollector.
// It exposes the kstat command result.
func NewZoneKstatExporter() (*ZoneKstatCollector, error) {
	return &ZoneKstatCollector{
		ZoneKstatCPUBaseline: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_cpu_baseline",
			Help: "A soft limit on the number of CPU cycles a hosted application can consume.",
		}, []string{"zonename"}),
		ZoneKstatCPUCap: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_cpu_cap",
			Help: "The maximum number of CPU cycles that are allocated to a zone.",
		}, []string{"zonename"}),
		ZoneKstatCPUMaxUsage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_cpu_maxusage",
			Help: "The maximum percentage of CPU used.",
		}, []string{"zonename"}),
		ZoneKstatCPUUsage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_cpu_usage",
			Help: "The current percentage of CPU used.",
		}, []string{"zonename"}),
		ZoneKstatMemCap: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_memory_cap_bytes",
			Help: "The physical memory limit in bytes.",
		}, []string{"zonename"}),
		ZoneKstatMemFree: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_memory_free_bytes",
			Help: "Free memory available in bytes.",
		}, []string{"zonename"}),
		ZoneKstatMemNover: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_memory_nover_total",
			Help: "The number of times the zone has gone over its cap.",
		}, []string{"zonename"}),
		ZoneKstatMemPagedOut: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_memory_pagedout_bytes",
			Help: "Total amount of memory that has been paged out when the zone has gone over its cap.",
		}, []string{"zonename"}),
		ZoneKstatMemRSS: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_memory_rss_bytes",
			Help: "Entire amount of allocated memory.",
		}, []string{"zonename"}),
		ZoneKstatNICCollisions: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_network_collisions",
			Help: "Entire amount of collisions.",
		}, []string{"zonename", "device"}),
		ZoneKstatNICIErrors: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_network_receive_errs_total",
			Help: "Received errors.",
		}, []string{"zonename", "device"}),
		ZoneKstatNICIPackets: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_network_receive_packets_total",
			Help: "Frames received successfully.",
		}, []string{"zonename", "device"}),
		ZoneKstatNICLinkState: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_network_link_state",
			Help: "Link state; 0 for down, 1 for up.",
		}, []string{"zonename", "device"}),
		ZoneKstatNICOBytes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_network_transmit_bytes_total",
			Help: "Bytes (octets) transmitted successfully.",
		}, []string{"zonename", "device"}),
		ZoneKstatNICOErrors: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_network_transmit_errs_total",
			Help: "Transmit errors.",
		}, []string{"zonename", "device"}),
		ZoneKstatNICOPackets: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_network_transmit_packets_total",
			Help: "Frames successfully transmitted.",
		}, []string{"zonename", "device"}),
		ZoneKstatNICRBytes: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_network_receive_bytes_total",
			Help: "Bytes (octets) received successfully.",
		}, []string{"zonename", "device"}),
		ZoneKstatSwapCap: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_memory_swap_cap_bytes",
			Help: "The SWAP limit in bytes.",
		}, []string{"zonename"}),
		ZoneKstatSwapFree: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_memory_swap_free_bytes",
			Help: "Free SWAP available in bytes.",
		}, []string{"zonename"}),
		ZoneKstatSwapUsed: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "smartos_memory_swap_used_bytes",
			Help: "Used SWAP in bytes.",
		}, []string{"zonename"}),
	}, nil
}

// Describe describes all the metrics.
func (e *ZoneKstatCollector) Describe(ch chan<- *prometheus.Desc) {
	e.ZoneKstatCPUBaseline.Describe(ch)
	e.ZoneKstatCPUCap.Describe(ch)
	e.ZoneKstatCPUMaxUsage.Describe(ch)
	e.ZoneKstatCPUUsage.Describe(ch)
	e.ZoneKstatMemCap.Describe(ch)
	e.ZoneKstatMemFree.Describe(ch)
	e.ZoneKstatMemNover.Describe(ch)
	e.ZoneKstatMemPagedOut.Describe(ch)
	e.ZoneKstatMemRSS.Describe(ch)
	e.ZoneKstatNICCollisions.Describe(ch)
	e.ZoneKstatNICIErrors.Describe(ch)
	e.ZoneKstatNICIPackets.Describe(ch)
	e.ZoneKstatNICLinkState.Describe(ch)
	e.ZoneKstatNICOBytes.Describe(ch)
	e.ZoneKstatNICOErrors.Describe(ch)
	e.ZoneKstatNICOPackets.Describe(ch)
	e.ZoneKstatNICRBytes.Describe(ch)
	e.ZoneKstatSwapCap.Describe(ch)
	e.ZoneKstatSwapFree.Describe(ch)
	e.ZoneKstatSwapUsed.Describe(ch)
}

// Collect fetches the stats.
func (e *ZoneKstatCollector) Collect(ch chan<- prometheus.Metric) {
	e.kstatCPUList()
	e.kstatMemList()
	e.kstatNICList()
	e.ZoneKstatCPUBaseline.Collect(ch)
	e.ZoneKstatCPUCap.Collect(ch)
	e.ZoneKstatCPUMaxUsage.Collect(ch)
	e.ZoneKstatCPUUsage.Collect(ch)
	e.ZoneKstatMemCap.Collect(ch)
	e.ZoneKstatMemFree.Collect(ch)
	e.ZoneKstatMemNover.Collect(ch)
	e.ZoneKstatMemPagedOut.Collect(ch)
	e.ZoneKstatMemRSS.Collect(ch)
	e.ZoneKstatNICCollisions.Collect(ch)
	e.ZoneKstatNICIErrors.Collect(ch)
	e.ZoneKstatNICIPackets.Collect(ch)
	e.ZoneKstatNICLinkState.Collect(ch)
	e.ZoneKstatNICOBytes.Collect(ch)
	e.ZoneKstatNICOErrors.Collect(ch)
	e.ZoneKstatNICOPackets.Collect(ch)
	e.ZoneKstatNICRBytes.Collect(ch)
	e.ZoneKstatSwapCap.Collect(ch)
	e.ZoneKstatSwapFree.Collect(ch)
	e.ZoneKstatSwapUsed.Collect(ch)
}

func (e *ZoneKstatCollector) kstatCPUList() {
	out, eerr := exec.Command("kstat", "-p", "-c", "zone_caps", "-n", "cpucaps_zone*").Output()
	if eerr != nil {
		log.Errorf("error on executing kstat: %v", eerr)
	}
	perr := e.parseKstatCPUListOutput(string(out))
	if perr != nil {
		log.Errorf("error on parsing kstat CPU list: %v", perr)
	}
}

func (e *ZoneKstatCollector) kstatMemList() {
	out, eerr := exec.Command("kstat", "-p", "-c", "zone_memory_cap").Output()
	if eerr != nil {
		log.Errorf("error on executing kstat: %v", eerr)
	}
	perr := e.parseKstatMemListOutput(string(out))
	if perr != nil {
		log.Errorf("error on parsing kstat Mem list: %v", perr)
	}
}

func (e *ZoneKstatCollector) kstatNICList() {
	out, eerr := exec.Command("kstat", "-p", "-m", "link").Output()
	if eerr != nil {
		log.Errorf("error on executing kstat: %v", eerr)
	}
	perr := e.parseKstatNICListOutput(string(out))
	if perr != nil {
		log.Errorf("error on parsing kstat NIC list: %v", perr)
	}
}

func (e *ZoneKstatCollector) parseKstatCPUListOutput(out string) error {
	// trim the label in order to obtain the type of metric
	r, _ := regexp.Compile(`(?m)^.*:.*:.*:`)

	outlines := strings.Split(out, "\n")
	l := len(outlines)
	m := make(map[string]string)
	for _, line := range outlines[1 : l-1] {
		parsedLine := strings.Fields(line)
		fullLabel := parsedLine[0]
		label := r.ReplaceAllString(fullLabel, "")
		// map label and values
		m[label] = parsedLine[1]
	}

	baseline, err := strconv.ParseFloat(m["baseline"], 64)
	if err != nil {
		return err
	}
	cap, err := strconv.ParseFloat(m["value"], 64)
	if err != nil {
		return err
	}
	maxUsage, err := strconv.ParseFloat(m["maxusage"], 64)
	if err != nil {
		return err
	}
	usage, err := strconv.ParseFloat(m["usage"], 64)
	if err != nil {
		return err
	}

	e.ZoneKstatCPUBaseline.With(prometheus.Labels{"zonename": m["zonename"]}).Set(baseline)
	e.ZoneKstatCPUCap.With(prometheus.Labels{"zonename": m["zonename"]}).Set(cap)
	e.ZoneKstatCPUMaxUsage.With(prometheus.Labels{"zonename": m["zonename"]}).Set(maxUsage)
	e.ZoneKstatCPUUsage.With(prometheus.Labels{"zonename": m["zonename"]}).Set(usage)

	return nil
}

func (e *ZoneKstatCollector) parseKstatMemListOutput(out string) error {
	// trim the label in order to obtain the type of metric
	r, _ := regexp.Compile(`(?m)^.*:.*:.*:`)

	outlines := strings.Split(out, "\n")
	l := len(outlines)
	m := make(map[string]string)
	for _, line := range outlines[1 : l-1] {
		parsedLine := strings.Fields(line)
		fullLabel := parsedLine[0]
		label := r.ReplaceAllString(fullLabel, "")
		// map label and values
		m[label] = parsedLine[1]
	}

	memCap, err := strconv.ParseFloat(m["physcap"], 64)
	if err != nil {
		return err
	}
	memNover, err := strconv.ParseFloat(m["nover"], 64)
	if err != nil {
		return err
	}
	memPagedOut, err := strconv.ParseFloat(m["pagedout"], 64)
	if err != nil {
		return err
	}
	memRSS, err := strconv.ParseFloat(m["rss"], 64)
	if err != nil {
		return err
	}
	memFree := memCap - memRSS

	swapCap, err := strconv.ParseFloat(m["swapcap"], 64)
	if err != nil {
		return err
	}
	swapUsed, err := strconv.ParseFloat(m["swap"], 64)
	if err != nil {
		return err
	}
	swapFree := memCap - memRSS

	e.ZoneKstatMemCap.With(prometheus.Labels{"zonename": m["zonename"]}).Set(memCap)
	e.ZoneKstatMemFree.With(prometheus.Labels{"zonename": m["zonename"]}).Set(memFree)
	e.ZoneKstatMemNover.With(prometheus.Labels{"zonename": m["zonename"]}).Set(memNover)
	e.ZoneKstatMemPagedOut.With(prometheus.Labels{"zonename": m["zonename"]}).Set(memPagedOut)
	e.ZoneKstatMemRSS.With(prometheus.Labels{"zonename": m["zonename"]}).Set(memRSS)

	e.ZoneKstatSwapCap.With(prometheus.Labels{"zonename": m["zonename"]}).Set(swapCap)
	e.ZoneKstatSwapFree.With(prometheus.Labels{"zonename": m["zonename"]}).Set(swapFree)
	e.ZoneKstatSwapUsed.With(prometheus.Labels{"zonename": m["zonename"]}).Set(swapUsed)

	return nil
}

func (e *ZoneKstatCollector) parseKstatNICListOutput(out string) error {
	// trim the label in order to obtain the type of metric and interface name
	r, _ := regexp.Compile(`(?m)^.+:(.+):`)

	outlines := strings.Split(out, "\n")
	l := len(outlines)
	m := make(map[ZoneKstatNIC]string)
	for _, line := range outlines[1 : l-1] {
		parsedLine := strings.Fields(line)
		fullLabel := parsedLine[0]
		ifname := r.FindStringSubmatch(fullLabel)[1]
		label := r.ReplaceAllString(fullLabel, "")
		// map struct label and values
		m[ZoneKstatNIC{ifname, label}] = parsedLine[1]
	}

	// populates the slice with the map
	var ZoneKstatNICkeys ZoneKstatNICs
	for k := range m {
		ZoneKstatNICkeys = append(ZoneKstatNICkeys, k)
	}

	for _, k := range ZoneKstatNICkeys {
		if k.ifLabel == "collisions" {
			collisions, err := strconv.ParseFloat(m[ZoneKstatNIC{k.ifName, "collisions"}], 64)
			if err != nil {
				return err
			}
			e.ZoneKstatNICCollisions.With(
				prometheus.Labels{"zonename": m[ZoneKstatNIC{k.ifName, "zonename"}], "device": k.ifName},
			).Set(collisions)
		}
		if k.ifLabel == "ierrors" {
			ierrors, err := strconv.ParseFloat(m[ZoneKstatNIC{k.ifName, "ierrors"}], 64)
			if err != nil {
				return err
			}
			e.ZoneKstatNICIErrors.With(
				prometheus.Labels{"zonename": m[ZoneKstatNIC{k.ifName, "zonename"}], "device": k.ifName},
			).Set(ierrors)
		}
		if k.ifLabel == "ipackets64" {
			ipackets64, err := strconv.ParseFloat(m[ZoneKstatNIC{k.ifName, "ipackets64"}], 64)
			if err != nil {
				return err
			}
			e.ZoneKstatNICIPackets.With(
				prometheus.Labels{"zonename": m[ZoneKstatNIC{k.ifName, "zonename"}], "device": k.ifName},
			).Set(ipackets64)
		}
		if k.ifLabel == "link_state" {
			linkState, err := strconv.ParseFloat(m[ZoneKstatNIC{k.ifName, "link_state"}], 64)
			if err != nil {
				return err
			}
			e.ZoneKstatNICLinkState.With(
				prometheus.Labels{"zonename": m[ZoneKstatNIC{k.ifName, "zonename"}], "device": k.ifName},
			).Set(linkState)
		}
		if k.ifLabel == "obytes64" {
			obytes64, err := strconv.ParseFloat(m[ZoneKstatNIC{k.ifName, "obytes64"}], 64)
			if err != nil {
				return err
			}
			e.ZoneKstatNICOBytes.With(
				prometheus.Labels{"zonename": m[ZoneKstatNIC{k.ifName, "zonename"}], "device": k.ifName},
			).Set(obytes64)
		}
		if k.ifLabel == "oerrors" {
			oerrors, err := strconv.ParseFloat(m[ZoneKstatNIC{k.ifName, "oerrors"}], 64)
			if err != nil {
				return err
			}
			e.ZoneKstatNICOErrors.With(
				prometheus.Labels{"zonename": m[ZoneKstatNIC{k.ifName, "zonename"}], "device": k.ifName},
			).Set(oerrors)
		}
		if k.ifLabel == "opackets64" {
			opackets64, err := strconv.ParseFloat(m[ZoneKstatNIC{k.ifName, "opackets64"}], 64)
			if err != nil {
				return err
			}
			e.ZoneKstatNICOPackets.With(
				prometheus.Labels{"zonename": m[ZoneKstatNIC{k.ifName, "zonename"}], "device": k.ifName},
			).Set(opackets64)
		}
		if k.ifLabel == "rbytes64" {
			rbytes64, err := strconv.ParseFloat(m[ZoneKstatNIC{k.ifName, "rbytes64"}], 64)
			if err != nil {
				return err
			}
			e.ZoneKstatNICRBytes.With(
				prometheus.Labels{"zonename": m[ZoneKstatNIC{k.ifName, "zonename"}], "device": k.ifName},
			).Set(rbytes64)
		}
	}

	return nil
}
