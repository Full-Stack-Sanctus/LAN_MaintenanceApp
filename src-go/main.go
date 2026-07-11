package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gosnmp/gosnmp"
)

var ttlRegex = regexp.MustCompile(`(?i)ttl[=\s](\d+)`)

type Device struct {
	IP          string `json:"ip"`
	MAC         string `json:"mac"`
	Status      string `json:"status"`
	TTL         int    `json:"ttl"`
	OS          string `json:"os"`
	SubnetMatch bool   `json:"subnetMatch"`
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

// 1. MANAGED: Querying Core Switch OID and extracting IP from OID suffix index
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

	var devices []Device
	// OID for ipNetToMediaPhysAddress contains the mapping table
	err := params.BulkWalk(".1.3.6.1.2.1.4.22.1.2", func(pdu gosnmp.SnmpPDU) error {
		if pdu.Value == nil {
			return nil
		}
		
		// Extract IP address from trailing OID elements (.1.sub.ip.ip.ip.ip)
		parts := strings.Split(strings.TrimPrefix(pdu.Name, "."), ".")
		if len(parts) < 4 {
			return nil
		}
		ipStr := strings.Join(parts[len(parts)-4:], ".")

		// Cast payload byte slice to formatted Hex string representation safely
		var macStr string
		if bytes, ok := pdu.Value.([]byte); ok {
			var hexParts []string
			for _, b := range bytes {
				hexParts = append(hexParts, fmt.Sprintf("%02x", b))
			}
			macStr = strings.Join(hexParts, ":")
		} else {
			macStr = "Unknown"
		}

		devices = append(devices, Device{
			IP:          ipStr,
			MAC:         macStr,
			Status:      "Verified",
			SubnetMatch: true,
			OS:          "Managed Network Node",
		})
		return nil
	})

	if err != nil {
		return []Device{}
	}
	return devices
}

// 2. UNMANAGED: Two-Phase Discovery Sequence (Blaster + Global Reaper Cache)
func scanUnmanaged(cidr string) []Device {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return []Device{{IP: "Error", Status: "Invalid CIDR"}}
	}

	const workerCount = 64
	jobs := make(chan string, 1024)
	resultsChan := make(chan Device, 1024)
	var wg sync.WaitGroup

	// Phase 1: Fire Workers concurrently to wake targets up and open TCP connections
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for targetIP := range jobs {
				ttl := getTTL(targetIP)
				ports := scanPorts(targetIP)

				// CATCH STEALTH HOSTS: Online if they reply to ping OR expose open listening services
				isOnline := ttl > 0 || len(ports) > 0
				if !isOnline {
					continue 
				}

				status := "Online"
				osGuess := detectOSAdvanced(ttl, ports)
				subnetMatch := checkSubnet(targetIP, cidr)

				resultsChan <- Device{
					IP:          targetIP,
					Status:      status,
					TTL:         ttl,
					OS:          osGuess,
					SubnetMatch: subnetMatch,
				}
			}
		}()
	}

	// Queue up subnet space blocks
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ipCopy := make(net.IP, len(ip))
		copy(ipCopy, ip)
		jobs <- ipCopy.String()
	}
	close(jobs)
	wg.Wait()
	close(resultsChan)

	// Collect intermediate records
	discoveredMap := make(map[string]Device)
	for d := range resultsChan {
		discoveredMap[d.IP] = d
	}

	// Phase 2: Wait briefly for kernel cache sync, dump full OS table lookup safely
	time.Sleep(50 * time.Millisecond)
	arpCache := readSystemARPCache()

	var finalDevices []Device
	for ipStr, device := range discoveredMap {
		if mac, exists := arpCache[ipStr]; exists {
			device.MAC = mac
		} else {
			device.MAC = "Unknown"
		}
		finalDevices = append(finalDevices, device)
	}

	return finalDevices
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func getTTL(ip string) int {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", "1", "-w", "800", ip)
	} else {
		cmd = exec.Command("ping", "-c", "1", "-W", "1", ip)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0
	}

	matches := ttlRegex.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return 0
	}

	ttl, _ := strconv.Atoi(matches[1])
	return ttl
}

// Unified Complete Table Reaper to extract MAC addresses without race conditions
func readSystemARPCache() map[string]string {
	cache := make(map[string]string)
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("arp", "-a")
	} else {
		cmd = exec.Command("ip", "neigh", "show")
	}

	output, err := cmd.Output()
	if err != nil {
		return cache
	}

	lines := strings.Split(string(output), "\n")
	ipRegex := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	macRegex := regexp.MustCompile(`([0-9a-fA-F]{2}[:-]){5}([0-9a-fA-F]{2})`)

	for _, line := range lines {
		foundIP := ipRegex.FindString(line)
		foundMAC := macRegex.FindString(line)

		if foundIP != "" && foundMAC != "" {
			standardizedMAC := strings.ReplaceAll(strings.ToLower(foundMAC), "-", ":")
			cache[foundIP] = standardizedMAC
		}
	}
	return cache
}

func scanPorts(ip string) map[int]bool {
	ports := []int{22, 80, 135, 139, 443, 445}
	results := make(map[int]bool)

	for _, port := range ports {
		address := fmt.Sprintf("%s:%d", ip, port)
		conn, err := net.DialTimeout("tcp", address, 250*time.Millisecond)
		if err == nil {
			results[port] = true
			conn.Close()
		}
	}
	return results
}

func detectOSAdvanced(ttl int, ports map[int]bool) string {
	if ports[445] || ports[135] || ports[139] {
		return "Windows Server/Workstation"
	}
	if ports[22] {
		return "Linux/Unix Environment"
	}
	if ports[80] || ports[443] {
		return "Embedded Web Device"
	}
	if ttl >= 120 {
		return "Windows OS (TTL heuristic)"
	} else if ttl >= 60 {
		return "Linux/Unix OS (TTL heuristic)"
	}
	return "Generic Network Endpoint"
}

func checkSubnet(ip string, cidr string) bool {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return network.Contains(net.ParseIP(ip))
}