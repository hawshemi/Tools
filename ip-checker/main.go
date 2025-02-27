package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

// Data structures
type ASNData struct {
	ASN    string `json:"asn"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Route  string `json:"route"`
	Type   string `json:"type"`
}

type ThreatData struct {
	IsTOR           bool `json:"is_tor"`
	IsProxy         bool `json:"is_proxy"`
	IsDatacenter    bool `json:"is_datacenter"`
	IsAnonymous     bool `json:"is_anonymous"`
	IsKnownAttacker bool `json:"is_known_attacker"`
	IsKnownAbuser   bool `json:"is_known_abuser"`
	IsThreat        bool `json:"is_threat"`
	IsBogon         bool `json:"is_bogon"`
	Blocklists      []struct {
		Name string `json:"name"`
		Site string `json:"site"`
		Type string `json:"type"`
	} `json:"blocklists"`
}

// Results struct to hold all collected data
type IPResults struct {
	ASNData          ASNData
	ThreatData       ThreatData
	BrowserLeaksData map[string]string
}

// HTTP client with timeout
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func main() {
	ip := flag.String("ip", "", "IP address to lookup")
	flag.Parse()

	if *ip == "" {
		logrus.Fatal("Please provide an IP address using the --ip flag")
	}

	if !isValidIP(*ip) {
		logrus.Fatalf("Invalid IP address: %s", *ip)
	}

	apiKey := os.Getenv("IPDATA_API_KEY")
	if apiKey == "" {
		logrus.Fatal("API key missing. Set IPDATA_API_KEY environment variable")
	}

	results := fetchIPData(*ip, apiKey)
	displayResults(results)
}

// fetchIPData fetches all data concurrently
func fetchIPData(ip, apiKey string) IPResults {
	var results IPResults
	var wg sync.WaitGroup
	wg.Add(3)

	// Fetch ASN data
	go func() {
		defer wg.Done()
		url := fmt.Sprintf("https://api.ipdata.co/%s/asn?api-key=%s", ip, apiKey)
		logrus.Info("Fetching ASN data...")
		results.ASNData = fetchJSON[ASNData](url)
	}()

	// Fetch Threat data
	go func() {
		defer wg.Done()
		url := fmt.Sprintf("https://api.ipdata.co/%s/threat?api-key=%s", ip, apiKey)
		logrus.Info("Fetching Threat data...")
		results.ThreatData = fetchJSON[ThreatData](url)
	}()

	// Fetch BrowserLeaks data
	go func() {
		defer wg.Done()
		url := fmt.Sprintf("https://browserleaks.com/ip/%s", ip)
		logrus.Info("Fetching BrowserLeaks data...")
		results.BrowserLeaksData = fetchBrowserLeaksData(url)
	}()

	wg.Wait()
	return results
}

// fetchJSON is a generic function to fetch and decode JSON
func fetchJSON[T any](url string) T {
	var result T
	resp, err := httpClient.Get(url)
	if err != nil {
		logrus.Fatalf("Failed to fetch %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		logrus.Fatalf("Rate limit exceeded. Retry after %s seconds", retryAfter)
	}

	if resp.StatusCode != http.StatusOK {
		logrus.Fatalf("Unexpected status code %d from %s", resp.StatusCode, url)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logrus.Fatalf("Failed to decode JSON from %s: %v", url, err)
	}
	return result
}

// fetchBrowserLeaksData scrapes data from browserleaks.com
func fetchBrowserLeaksData(url string) map[string]string {
	resp, err := httpClient.Get(url)
	if err != nil {
		logrus.Fatalf("Failed to fetch BrowserLeaks data: %v", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logrus.Fatalf("Failed to parse BrowserLeaks HTML: %v", err)
	}

	fields := []string{"Country", "ISP", "Organization", "Usage Type"}
	data := make(map[string]string, len(fields))
	for _, field := range fields {
		data[field] = doc.Find(fmt.Sprintf("td:contains('%s')", field)).Next().Text()
		if data[field] == "" {
			data[field] = "N/A"
		}
	}
	return data
}

// displayResults shows all collected data in a tabulated format
func displayResults(results IPResults) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "\nCategory\tValue")

	// ASN Data
	fmt.Fprintf(w, "ASN\t%s\n", results.ASNData.ASN)
	fmt.Fprintf(w, "ASN Name\t%s\n", results.ASNData.Name)
	fmt.Fprintf(w, "ASN Type\t%s\n", results.ASNData.Type)

	// BrowserLeaks Data
	for key, value := range results.BrowserLeaksData {
		fmt.Fprintf(w, "%s\t%s\n", key, value)
	}

	// Threat Data
	threatFields := map[string]bool{
		"Is TOR":            results.ThreatData.IsTOR,
		"Is Proxy":          results.ThreatData.IsProxy,
		"Is Datacenter":     results.ThreatData.IsDatacenter,
		"Is Anonymous":      results.ThreatData.IsAnonymous,
		"Is Known Attacker": results.ThreatData.IsKnownAttacker,
		"Is Known Abuser":   results.ThreatData.IsKnownAbuser,
		"Is Threat":         results.ThreatData.IsThreat,
		"Is Bogon":          results.ThreatData.IsBogon,
	}
	for field, value := range threatFields {
		fmt.Fprintf(w, "%s\t%t\n", field, value)
	}

	// Blocklists
	if len(results.ThreatData.Blocklists) > 0 {
		fmt.Fprintln(w, "Blocklists:")
		for i, b := range results.ThreatData.Blocklists {
			fmt.Fprintf(w, "Blocklist %d Name\t%s\n", i+1, b.Name)
			fmt.Fprintf(w, "Site\t%s\n", b.Site)
			fmt.Fprintf(w, "Type\t%s\n", b.Type)
			if i < len(results.ThreatData.Blocklists)-1 {
				fmt.Fprintln(w, "---\t---")
			}
		}
	} else {
		fmt.Fprintln(w, "Blocklists\tNone")
	}

	w.Flush()
}

// isValidIP validates IP address format
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}
