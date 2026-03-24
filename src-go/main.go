package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"sync"
	"time"
	"strings"
	
	"os/exec"
	"regexp"
	"strconv"

	"github.com/gosnmp/gosnmp"
)

var (
	ttlRegex = regexp.MustCompile(`ttl[=\s](\d+)`)
	macRegex = regexp.MustCompile(`lladdr\s+([0-9a-fA-F:]+)`)
)

type Device struct {
	IP           string `json:"ip"`
	MAC          string `json:"mac"`
	Status       string `json:"status"`
	TTL          int    `json:"ttl"`
	OS           string `json:"os"`
	SubnetMatch  bool   `json:"subnetMatch"`
}

type NetworkReport struct {
	Devices     []Device `json:"devices"`
	ScanMethod  string   `json:"scanMethod"`
	Subnet      string   `json:"subnet"`
	Timestamp   string   `json:"timestamp"`
	Performance string   `json:"performance"`
}

func main() {
	target := flag.String("target", "192.168.1.0/24", "CIDR Subnet or Switch IP")
	community := flag.String("community", "", "SNMP Community String")
	flag.Parse()

	report := NetworkReport{
		Subnet:    *target,
		Timestamp: time.Now().Format(time.RFC1123),
	}

	if *community != "" {
		report.Devices = scanManaged(*target, *community)
		report.ScanMethod = "SNMP (Core Switch)"
	} else {
		report.Devices = scanUnmanaged(*target)
		report.ScanMethod = "CIDR ARP Sweep"
	}

	report.Performance = "Optimal"
	data, _ := json.Marshal(report)
	fmt.Println(string(data))
}

// 1. MANAGED: Querying the Core Switch OID Table
func scanManaged(target, community string) []Device {
	params := &gosnmp.GoSNMP{
		Target:    target,
		Port:      161,
		Community: community,
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(2) * time.Second,
	}
	if err := params.Connect(); err != nil {
		return []Device{}
	}
	defer params.Conn.Close()

	devices := []Device{}
	// OID for ipNetToMediaPhysAddress (Returns MACs and IPs from Switch Cache)
	err := params.BulkWalk(".1.3.6.1.2.1.4.22.1.2", func(pdu gosnmp.SnmpPDU) error {
		devices = append(devices, Device{IP: "From Cache", MAC: fmt.Sprintf("%x", pdu.Value), Status: "Verified"})
		return nil
	})
	if err != nil {
		return []Device{}
	}
	return devices
}

// 2. UNMANAGED: Concurrent CIDR Probing
func scanUnmanaged(cidr string) []Device {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return []Device{{IP: "Error", Status: "Invalid CIDR"}}
	}

	const workerCount = 50 // 🔥 Tune this (30–100 depending on system)

	jobs := make(chan string, 512)
	resultsChan := make(chan Device, 512)

	var wg sync.WaitGroup

	// 🔥 Worker Pool
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for targetIP := range jobs {

				ttl := getTTL(targetIP)
				if ttl == 0 {
					continue
				}

				mac := getMAC(targetIP)
				os := detectOS(ttl)
				subnetMatch := checkSubnet(targetIP, cidr)

				resultsChan <- Device{
					IP: targetIP,
					MAC: mac,
					Status: "Online",
					TTL: ttl,
					OS: os,
					SubnetMatch: subnetMatch,
				}
			}
		}()
	}

	// 🔥 Feed jobs
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		jobs <- ip.String()
	}
	close(jobs)

	// 🔥 Close results when done
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	results := []Device{}
	for d := range resultsChan {
		results = append(results, d)
	}

	return results
}

// ICMP TTL
func getTTL(ip string) int {
	cmd := exec.Command("ping", "-c", "1", "-W", "1", ip)

	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	matches := ttlRegex.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return 0
	}

	ttl, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}

	return ttl
}

//ARP Lookup
func getMAC(ip string) string {
	cmd := exec.Command("ip", "neigh", "show", ip)

	output, err := cmd.Output()
	if err != nil {
		return "Unknown"
	}

	matches := macRegex.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return "Unknown"
	}

	return matches[1]
}

// OS FINGERPRINT (TTL Heuristic)
func detectOS(ttl int) string {
	if ttl >= 120 {
		return "Windows"
	} else if ttl >= 60 {
		return "Linux/Unix"
	} else {
		return "Network Device"
	}
}

// Subnet check
func checkSubnet(ip string, cidr string) bool {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return network.Contains(net.ParseIP(ip))
}