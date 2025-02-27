package main

import (
	"bufio"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Provider struct {
	IP   string
	Name string
}

type TestResult struct {
	Name           string
	IP             string
	AvgPing        string
	AvgResolveTime string
}

const (
	failMessage    = "FAIL"
	pingCount      = 4
	resolveCount   = 4
	timeoutSeconds = 2
)

var (
	providersV4 = []Provider{
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

	providersV6 = []Provider{
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

	domains = []string{
		"www.google.com",
		"www.speedtest.net",
		"i.instagram.com",
	}
)

func getAveragePing(ip string) string {
	cmd := exec.Command("ping", "-c", strconv.Itoa(pingCount), "-W", strconv.Itoa(timeoutSeconds), ip)
	output, err := cmd.Output()
	if err != nil {
		return failMessage
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "avg") {
			parts := strings.Split(line, "/")
			if len(parts) >= 5 {
				return parts[4] + " ms"
			}
		}
	}
	return failMessage
}

func resolveDomain(domain, server string) float64 {
	var totalTime float64
	for i := 0; i < resolveCount; i++ {
		uniqueDomain := fmt.Sprintf("%x.%s", time.Now().UnixNano(), domain)
		cmd := exec.Command("dig", "@"+server, uniqueDomain, "+time="+strconv.Itoa(timeoutSeconds), "+tries=1")
		output, err := cmd.Output()
		if err != nil {
			return math.NaN()
		}

		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "Query time") {
				parts := strings.Fields(line)
				if len(parts) >= 4 {
					if timeInt, err := strconv.Atoi(parts[3]); err == nil {
						totalTime += float64(timeInt)
					}
				}
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return totalTime / float64(resolveCount)
}

func printTable(results []TestResult) {
	fmt.Println("+---------------------------------+-----------------------+---------------+------------------------+")
	fmt.Println("| Provider Name                   | IP Address            | Avg Ping (ms) | Avg Resolve Time (ms) |")
	fmt.Println("+---------------------------------+-----------------------+---------------+------------------------+")
	
	for _, result := range results {
		fmt.Printf("| %-31s | %-21s | %-13s | %-21s |\n",
			result.Name, result.IP, result.AvgPing, result.AvgResolveTime)
	}
	
	fmt.Println("+---------------------------------+-----------------------+---------------+------------------------+")
}

func isIPv6Enabled() bool {
	cmd := exec.Command("ping6", "-c", "1", "-W", "1", "google.com")
	return cmd.Run() == nil
}

func testProvider(provider Provider, resultChan chan<- TestResult, wg *sync.WaitGroup) {
	defer wg.Done()

	// Run ping test
	avgPing := getAveragePing(provider.IP)

	// Run DNS resolution tests concurrently
	var resolveWG sync.WaitGroup
	resolveTimes := make([]float64, len(domains))
	for i, domain := range domains {
		resolveWG.Add(1)
		go func(idx int, d string) {
			defer resolveWG.Done()
			resolveTimes[idx] = resolveDomain(d, provider.IP)
		}(i, domain)
	}
	resolveWG.Wait()

	// Calculate average resolve time
	var totalResolveTime float64
	var validResults int
	for _, time := range resolveTimes {
		if !math.IsNaN(time) {
			totalResolveTime += time
			validResults++
		}
	}

	result := TestResult{
		Name:    provider.Name,
		IP:      provider.IP,
		AvgPing: avgPing,
	}

	if validResults == 0 {
		result.AvgResolveTime = failMessage
	} else {
		avgResolveTime := totalResolveTime / float64(validResults)
		result.AvgResolveTime = fmt.Sprintf("%.1f ms", avgResolveTime)
	}

	resultChan <- result
}

func main() {
	// Create channels and wait group for concurrent testing
	resultChan := make(chan TestResult, len(providersV4)+len(providersV6))
	var wg sync.WaitGroup
	var results []TestResult

	// Test IPv4 providers
	for _, provider := range providersV4 {
		wg.Add(1)
		go testProvider(provider, resultChan, &wg)
	}

	// Test IPv6 providers if enabled
	if isIPv6Enabled() {
		for _, provider := range providersV6 {
			wg.Add(1)
			go testProvider(provider, resultChan, &wg)
		}
	}

	// Close result channel when all tests are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for result := range resultChan {
		results = append(results, result)
	}

	// Print results
	printTable(results)
}
