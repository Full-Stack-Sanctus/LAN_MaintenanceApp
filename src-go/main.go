package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/gosnmp/gosnmp"
)

type Device struct {
	IP          string `json:"ip"`
	MAC         string `json:"mac"`
	Status      string `json:"status"`
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
	target := flag.String("target", "192.168.1.0/24", "Target CIDR Subnet range")
	community := flag.String("community", "", "SNMP Community Key string")
	flag.Parse()

	report := NetworkReport{
		Subnet:    *target,
		Timestamp: time.Now().Format(time.RFC1123),
	}

	if *community != "" {
		report.Devices = scanManaged(*target, *community)
		report.ScanMethod = "L2 Infrastructure (SNMP Core Table)"
	} else {
		report.Devices = scanUnmanagedLayer2(*target)
		report.ScanMethod = "Layer 2 Raw ARP Broadcast Sweep"
	}

	report.Performance = "Production Optimized (Zero-OS Reliance)"
	data, _ := json.Marshal(report)
	fmt.Println(string(data))
}

// 1. MANAGED CORE ENGINE: Direct Hardware Bridge via SNMP
func scanManaged(target, community string) []Device {
	params := &gosnmp.GoSNMP{
		Target:    target,
		Port:      161,
		Community: community,
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(3) * time.Second,
	}

	if err := params.Connect(); err != nil {
		return []Device{}
	}

	defer params.Conn.Close()

	var devices []Device

	// OID mapping for ipNetToMediaPhysAddress table
	err := params.BulkWalk(".1.3.6.1.2.1.4.22.1.2", func(pdu gosnmp.SnmpPDU) error {
		if pdu.Value == nil {
			return nil
		}

		parts := strings.Split(strings.TrimPrefix(pdu.Name, "."), ".")
		if len(parts) < 4 {
			return nil
		}

		ipStr := strings.Join(parts[len(parts)-4:], ".")

		var macStr string

		if bytesVal, ok := pdu.Value.([]byte); ok {
			var hexParts []string

			for _, b := range bytesVal {
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

// 2. UNMANAGED CORE ENGINE: Raw Layer 2 Direct-to-NIC ARP Injector & Listener
func scanUnmanagedLayer2(cidr string) []Device {
	_, ipnet, err := net.ParseCIDR(cidr)

	if err != nil {
		return []Device{{IP: "Error", Status: "Invalid CIDR Target"}}
	}

	// Step A: Find the correct local interface mapping out to this destination route target
	mapping, err := findActiveInterface(ipnet)

	if err != nil {
		return []Device{{IP: "Interface Error", Status: err.Error()}}
	}

	iface := mapping.Interface
	localIP := mapping.LocalIP

	fmt.Printf(
		"Windows Interface: %s | IfIndex: %d | Local IP: %s | MAC: %s\n",
		iface.Name,
		iface.Index,
		localIP,
		iface.HardwareAddr,
	)

	fmt.Printf(
		"Npcap Device: %s | Description: %s\n",
		mapping.PcapDevice.Name,
		mapping.PcapDevice.Description,
	)

	// Step B: Initialize Npcap using the actual Npcap device name,
	// NOT the Windows friendly adapter name.
	handle, err := pcap.OpenLive(
		mapping.PcapDevice.Name,
		65536,
		true,
		pcap.BlockForever,
	)

	if err != nil {
		return []Device{{IP: "Pcap Error", Status: err.Error()}}
	}

	if err != nil {
		return []Device{{IP: "Pcap Error", Status: err.Error()}}
	}

	defer handle.Close()

	// Enforce strict kernel-level BPF filter to parse *only* incoming ARP replies meant for us
	// filter := fmt.Sprintf("arp and ether dst %s", iface.HardwareAddr.String())

	if err := handle.SetBPFFilter(filter); err != nil {
		return []Device{{IP: "BPF Error", Status: err.Error()}}
	}

	discoveredDevices := &sync.Map{}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2500*time.Millisecond,
	)
	defer cancel()

	// Step C: Spin up background Listener worker parsing raw physical responses
	var wgListener sync.WaitGroup

	wgListener.Add(1)

	go func() {
		defer wgListener.Done()

		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		inChan := packetSource.Packets()

		for {
			select {
			case <-ctx.Done():
				return

			case packet, ok := <-inChan:
				if !ok {
					return
				}

				arpLayer := packet.Layer(layers.LayerTypeARP)

				if arpLayer == nil {
					continue
				}

				arpPad := arpLayer.(*layers.ARP)

				// Validate it's an ARP Reply operation
				if arpPad.Operation != layers.ARPReply {
					continue
				}

				foundIP := net.IP(arpPad.SourceProtAddress).String()
				foundMAC := net.HardwareAddr(arpPad.SourceHwAddress).String()

				discoveredDevices.Store(foundIP, Device{
					IP:          foundIP,
					MAC:         foundMAC,
					Status:      "Online",
					OS:          "Hardware Validated (L2 Reply)",
					SubnetMatch: true,
				})
			}
		}
	}()

	// Step D: Spin up concurrent Transmitter pool injecting raw ARP frames directly into the wire
	var wgSender sync.WaitGroup
	ipsToScan := make(chan net.IP, 256)

	// Spawning standard balanced transmitter lanes
	for w := 0; w < 16; w++ {
		wgSender.Add(1)

		go func() {
			defer wgSender.Done()

			for targetIP := range ipsToScan {
				_ = writeARPFrame(handle, iface, localIP, targetIP)
				time.Sleep(2 * time.Millisecond) // Throttled to prevent local buffering starvation
			}
		}()
	}

	// Stream targets over execution loop
	for ip := ipnet.IP.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		if ip.Equal(ipnet.IP) || ip.Equal(lastIP(ipnet)) {
			continue // Skip network index boundary limits (.0 and .255)
		}

		ipCopy := make(net.IP, len(ip))
		copy(ipCopy, ip)

		ipsToScan <- ipCopy
	}

	close(ipsToScan)
	wgSender.Wait()

	// Allow network buffer cooldown period to let trailing frames return safely
	time.Sleep(600 * time.Millisecond)

	cancel()
	wgListener.Wait()

	// Assemble dynamic map payloads into formal clean collection slices
	var finalDevices []Device

	discoveredDevices.Range(func(key, value interface{}) bool {
		finalDevices = append(finalDevices, value.(Device))
		return true
	})

	return finalDevices
}

func writeARPFrame(
	handle *pcap.Handle,
	iface *net.Interface,
	srcIP net.IP,
	dstIP net.IP,
) error {
	ethLayer := &layers.Ethernet{
		SrcMAC:       iface.HardwareAddr,
		DstMAC:       net.HardwareAddr{
			0xff,
			0xff,
			0xff,
			0xff,
			0xff,
			0xff,
		}, // Broadcast target address Frame block
		EthernetType: layers.EthernetTypeARP,
	}

	arpLayer := &layers.ARP{
		AddrType:         layers.LinkTypeEthernet,
		Protocol:         layers.EthernetTypeIPv4, // Fixed: ProtoType -> Protocol
		HwAddressSize:    6,
		ProtAddressSize:  4,
		Operation:        layers.ARPRequest, // Fixed: ARPOpRequest -> ARPRequest
		SourceHwAddress:  []byte(iface.HardwareAddr),
		SourceProtAddress: []byte(srcIP.To4()),
		DstHwAddress:     []byte{
			0x00,
			0x00,
			0x00,
			0x00,
			0x00,
			0x00,
		},
		DstProtAddress: []byte(dstIP.To4()),
	}

	buf := gopacket.NewSerializeBuffer()

	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	if err := gopacket.SerializeLayers(
		buf,
		opts,
		ethLayer,
		arpLayer,
	); err != nil {
		return err
	}

	return handle.WritePacketData(buf.Bytes())
}

type InterfaceMapping struct {
	Interface  *net.Interface
	LocalIP    net.IP
	PcapDevice pcap.Interface
}

func findActiveInterface(targetNet *net.IPNet) (*InterfaceMapping, error) {
	// Get operating-system network interfaces.
	ifaces, err := net.Interfaces()

	if err != nil {
		return nil, fmt.Errorf(
			"failed to enumerate network interfaces: %w",
			err,
		)
	}

	// Get Npcap interfaces.
	pcapDevices, err := pcap.FindAllDevs()

	if err != nil {
		return nil, fmt.Errorf(
			"failed to enumerate Npcap devices: %w",
			err,
		)
	}

	for _, iface := range ifaces {

		// Ignore interfaces that are down.
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Ignore loopback.
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()

		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)

			if !ok {
				continue
			}

			ip4 := ipNet.IP.To4()

			if ip4 == nil {
				continue
			}

			// The target must be directly reachable through
			// this local interface for an ARP scan.
			if !targetNet.Contains(ip4) {
				continue
			}

			// Now locate the corresponding Npcap device.
			for _, pcapDevice := range pcapDevices {

				// First try matching by IP address.
				for _, pcapAddr := range pcapDevice.Addresses {
					pcapIP := pcapAddr.IP.To4()

					if pcapIP == nil {
						continue
					}

					if pcapIP.Equal(ip4) {
						if pcapDevice.Name == "" {
							continue
						}

						return &InterfaceMapping{
							Interface:  &iface,
							LocalIP:    ip4,
							PcapDevice: pcapDevice,
						}, nil
					}
				}

				// Some Npcap versions/adapters may not expose
				// the address exactly as expected. Fall back
				// to matching the adapter description.
				if strings.EqualFold(
					strings.TrimSpace(pcapDevice.Description),
					strings.TrimSpace(iface.Name),
				) {
					if pcapDevice.Name == "" {
						continue
					}

					return &InterfaceMapping{
						Interface:  &iface,
						LocalIP:    ip4,
						PcapDevice: pcapDevice,
					}, nil
				}

				if strings.Contains(
					strings.ToLower(pcapDevice.Description),
					strings.ToLower(iface.Name),
				) {
					if pcapDevice.Name == "" {
						continue
					}

					return &InterfaceMapping{
						Interface:  &iface,
						LocalIP:    ip4,
						PcapDevice: pcapDevice,
					}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf(
		"no operational Npcap interface found for target subnet %s",
		targetNet.String(),
	)
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++

		if ip[j] > 0 {
			break
		}
	}
}

func lastIP(ipnet *net.IPNet) net.IP {
	var mask uint32

	if len(ipnet.Mask) == 4 {
		mask = binary.BigEndian.Uint32(ipnet.Mask)
	} else {
		return nil
	}

	num := binary.BigEndian.Uint32(ipnet.IP.To4())
	invMask := ^mask
	last := num | invMask

	res := make(net.IP, 4)

	binary.BigEndian.PutUint32(res, last)

	return res
}