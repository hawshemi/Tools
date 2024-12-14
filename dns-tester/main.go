package main

import (
	"bufio"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Provider struct {
	IP   string
	Name string
}

const failMessage = "FAIL"

var providersV4 = []Provider{
	{"1.1.1.1", "Cloudflare (v4)"},
	{"1.1.1.2", "Cloudflare-Security (v4)"},
	{"8.8.8.8", "Google (v4)"},
	{"9.9.9.9", "Quad9 (v4)"},
	{"209.244.0.3", "Level3 (v4)"},
	{"94.140.14.14", "Adguard (v4)"},
	{"193.110.81.0", "Dns0.eu (v4)"},
	{"76.76.2.2", "ControlD (v4)"},
	{"95.85.95.85", "GcoreDNS (v4)"},
	{"185.228.168.9", "CleanBrowsing-Security (v4)"},
	{"208.67.222.222", "OpenDNS (v4)"},
	{"77.88.8.8", "Yandex (v4)"},
	{"77.88.8.88", "Yandex-Safe (v4)"},
	{"64.6.64.6", "UltraDNS (v4)"},
	{"156.154.70.2", "UltraDNS-ThreatProtection (v4)"},
}

var providersV6 = []Provider{
	{"2606:4700:4700::1111", "Cloudflare (v6)"},
	{"2606:4700:4700::1112", "Cloudflare-Security (v6)"},
	{"2001:4860:4860::8888", "Google (v6)"},
	{"2620:fe::fe", "Quad9 (v6)"},
	{"2a10:50c0::ad1:ff", "Adguard (v6)"},
	{"2a0f:fc80::", "Dns0.eu (v6)"},
	{"2606:1a40::2", "ControlD (v6)"},
	{"2a03:90c0:999d::1", "GcoreDNS (v6)"},
	{"2a0d:2a00:1::2", "CleanBrowsing-Security (v6)"},
	{"2a02:6b8::feed:0ff", "Yandex (v6)"},
	{"2a02:6b8::feed:bad", "Yandex-Safe (v6)"},
	{"2620:74:1b::1:1", "UltraDNS (v6)"},
	{"2610:a1:1018::2", "UltraDNS-ThreatProtection (v6)"},
}

var domains = []string{"www.google.com", "www.speedtest.net", "i.instagram.com"}

func getAveragePing(ip string) string {
	cmd := exec.Command("ping", "-c", "4", ip)
	output, err := cmd.Output()
	if err != nil {
		return failMessage
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "avg") {
			parts := strings.Split(line, "/")
			if len(parts) >= 5 {
				return parts[4] + " ms"
			}
		}
	}
	return failMessage
}

func resolveDomain(domain, server string) string {
	var totalTime int
	for i := 0; i < 4; i++ {
		uniqueDomain := fmt.Sprintf("%x.%s", time.Now().UnixNano(), domain)
		cmd := exec.Command("dig", "@"+server, uniqueDomain, "+time=2", "+tries=1")
		output, err := cmd.Output()
		if err != nil {
			return failMessage
		}

		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "Query time") {
				parts := strings.Fields(line)
				if len(parts) >= 4 {
					timeInt, err := strconv.Atoi(parts[3])
					if err != nil {
						return failMessage
					}
					totalTime += timeInt
				}
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	avgTime := float64(totalTime) / 4.0
	if math.IsNaN(avgTime) {
		return failMessage
	}
	return fmt.Sprintf("%.1f ms", avgTime)
}

func printHeader() {
	fmt.Println("+---------------------------------+-----------------------+---------------+------------------------+")
	fmt.Println("| Provider Name                   | IP Address            | Avg Ping (ms) | Avg Resolve Time (ms) |")
	fmt.Println("+---------------------------------+-----------------------+---------------+------------------------+")
}

func printFooter() {
	fmt.Println("+---------------------------------+-----------------------+---------------+------------------------+")
}

func printRow(name, ip, avgPing, avgResolveTime string) {
	fmt.Printf("| %-31s | %-21s | %-13s | %-21s |\n", name, ip, avgPing, avgResolveTime)
}

func isIPv6Enabled() bool {
	cmd := exec.Command("ping6", "-c", "1", "google.com")
	if err := cmd.Run(); err != nil {
		return false
	}
	cmd = exec.Command("ip", "-6", "route", "show", "default")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func testProvider(provider Provider) {
	avgPing := getAveragePing(provider.IP)
	totalResolveTime := 0.0
	fail := false

	for _, domain := range domains {
		resolveTime := resolveDomain(domain, provider.IP)
		if resolveTime == failMessage {
			fail = true
			break
		}
		resolveTimeFloat, err := strconv.ParseFloat(strings.Fields(resolveTime)[0], 64)
		if err != nil {
			fail = true
			break
		}
		totalResolveTime += resolveTimeFloat
	}

	if fail {
		printRow(provider.Name, provider.IP, avgPing, failMessage)
	} else {
		avgResolveTime := totalResolveTime / float64(len(domains))
		printRow(provider.Name, provider.IP, avgPing, fmt.Sprintf("%.1f ms", avgResolveTime))
	}
}

func main() {
	printHeader()

	// Test IPv4 providers
	for _, provider := range providersV4 {
		testProvider(provider)
	}

	// Test IPv6 providers if enabled
	if isIPv6Enabled() {
		for _, provider := range providersV6 {
			testProvider(provider)
		}
	} else {
		fmt.Println("IPv6 is not enabled or not available. Skipping IPv6 tests.")
	}

	printFooter()
}
