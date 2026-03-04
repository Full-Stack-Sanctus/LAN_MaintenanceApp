package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
)

type NetworkReport struct {
	ActiveSubnets []string `json:"activeSubnets"`
	AlienIPs      []string `json:"alienIPs"`
	Performance   string   `json:"performance"`
	Suggestions   []string `json:"suggestions"`
}

func main() {
	// 1. Setup CLI flags so the UI can send parameters
	subnetFlag := flag.String("subnet", "192.168.1.0/24", "The subnet to scan")
	flag.Parse()

	report := analyzeNetwork(*subnetFlag)

	// 2. Format to JSON for the Frontend to consume
	data, err := json.Marshal(report)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 3. Print to Standard Out (Rust will capture this)
	fmt.Println(string(data))
}

func analyzeNetwork(mainSubnet string) NetworkReport {
	_, network, err := net.ParseCIDR(mainSubnet)
	report := NetworkReport{
		ActiveSubnets: []string{mainSubnet},
		AlienIPs:      []string{},
		Suggestions:   []string{},
	}

	if err != nil {
		report.Performance = "Error: Invalid Subnet"
		return report
	}

	// Mocking found IPs (In production, use gosnmp to query your switch)
	foundIPs := []string{"192.168.1.5", "10.0.0.50", "192.168.1.22"}

	for _, ipStr := range foundIPs {
		parsedIP := net.ParseIP(ipStr)
		if !network.Contains(parsedIP) {
			report.AlienIPs = append(report.AlienIPs, ipStr)
		}
	}

	// Logic for performance
	latency := 45 
	if len(report.AlienIPs) > 0 {
		report.Performance = "Degraded"
		report.Suggestions = append(report.Suggestions, "Rogue devices found. Verify VLAN isolation.")
	} else if latency > 30 {
		report.Performance = "Warning"
		report.Suggestions = append(report.Suggestions, "High latency. Check for SFP errors.")
	} else {
		report.Performance = "Optimal"
	}

	return report
}