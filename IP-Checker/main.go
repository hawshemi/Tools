package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

// Struct to hold the ipdata.co response for ASN data
type ASNData struct {
	ASN    string `json:"asn"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Route  string `json:"route"`
	Type   string `json:"type"`
}

// Struct to hold the ipdata.co response for Threat data
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

func main() {
	// Define the --ip flag to accept an IP address from the user
	ip := flag.String("ip", "", "IP address to lookup")
	flag.Parse()

	// Ensure that the IP flag is provided
	if *ip == "" {
		logrus.Fatalf("Please provide an IP address using the --ip flag")
	}

	// Validate IP Address
	if !isValidIP(*ip) {
		logrus.Fatalf("Invalid IP address: %s", *ip)
	}

	// Get the API key from the environment variable
	apiKey := os.Getenv("IPDATA_API_KEY")
	if apiKey == "" {
		logrus.Fatal("API key is missing. Set it via the IPDATA_API_KEY environment variable.")
	}

	// Fetch ASN data
	asnURL := fmt.Sprintf("https://api.ipdata.co/%s/asn?api-key=%s", *ip, apiKey)
	logrus.Info("Fetching ASN Data from ipdata.co.")
	asnData := fetchASNData(asnURL)

	// Fetch Threat data
	threatURL := fmt.Sprintf("https://api.ipdata.co/%s/threat?api-key=%s", *ip, apiKey)
	logrus.Info("Fetching Threat Data from ipdata.co.")
	threatData := fetchThreatData(threatURL)

	// Scrape and display data from browserleaks.com
	browserLeaksURL := fmt.Sprintf("https://browserleaks.com/ip/%s", *ip)
	logrus.Info("Fetching data from browserleaks.com.")
	browserLeaksData := fetchBrowserLeaksData(browserLeaksURL)

	// Display the results in a nice table format
	displayResults(asnData, threatData, browserLeaksData)
}

// Create a reusable HTTP client with a timeout
var httpClient = &http.Client{
	Timeout: 10 * time.Second, // 10 seconds timeout for all HTTP requests
}

// Function to make a GET request and return the response
func fetchJSON(url string, target interface{}) {
	resp, err := httpClient.Get(url)
	if err != nil {
		logrus.Fatalf("Failed to fetch data from %s: %v", url, err)
	}
	defer resp.Body.Close()

	// Handle rate limiting
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		logrus.Fatalf("Rate limit exceeded. Try again after %s seconds.", retryAfter)
	}

	// Check for non-OK response status
	if resp.StatusCode != http.StatusOK {
		logrus.Fatalf("Error: received status code %d from %s", resp.StatusCode, url)
	}

	// Decode JSON response into target struct
	err = json.NewDecoder(resp.Body).Decode(target)
	if err != nil {
		logrus.Fatalf("Failed to decode JSON response from %s: %v", url, err)
	}
}

// Fetch ASN data
func fetchASNData(url string) ASNData {
	var data ASNData
	fetchJSON(url, &data)
	return data
}

// Fetch Threat data
func fetchThreatData(url string) ThreatData {
	var data ThreatData
	fetchJSON(url, &data)
	return data
}

// Function to extract a specific field from the HTML document
func extractField(doc *goquery.Document, field string) string {
	return doc.Find(fmt.Sprintf("td:contains('%s')", field)).Next().Text()
}

// Fetch and parse data from browserleaks.com
func fetchBrowserLeaksData(url string) map[string]string {
	data := make(map[string]string)

	resp, err := httpClient.Get(url)
	if err != nil {
		logrus.Fatalf("Failed to fetch data from %s: %v", url, err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logrus.Fatalf("Failed to parse HTML from browserleaks.com: %v", err)
	}

	data["Country"] = extractField(doc, "Country")
	data["ISP"] = extractField(doc, "ISP")
	data["Organization"] = extractField(doc, "Organization")
	data["Usage Type"] = extractField(doc, "Usage Type")

	// Set default values if necessary
	for key, value := range data {
		if value == "" {
			data[key] = "N/A"
		}
	}

	return data
}

// Display formatted data in a table
func displayTable(w *tabwriter.Writer, category, value string) {
	fmt.Fprintf(w, "%s\t%s\t\n", category, value)
}

func displayResults(asnData ASNData, threatData ThreatData, browserLeaksData map[string]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)

	fmt.Fprintln(w, "\nCategory\tValue\t")

	// ASN Data
	displayTable(w, "ASN", asnData.ASN)
	displayTable(w, "ASN Name", asnData.Name)

	// Browserleaks Data
	for key, value := range browserLeaksData {
		displayTable(w, key, value)
	}

	displayTable(w, "ASN Type", asnData.Type)

	// Threat Data
	displayTable(w, "Is TOR", fmt.Sprintf("%t", threatData.IsTOR))
	displayTable(w, "Is Proxy", fmt.Sprintf("%t", threatData.IsProxy))
	displayTable(w, "Is Datacenter", fmt.Sprintf("%t", threatData.IsDatacenter))
	displayTable(w, "Is Anonymous", fmt.Sprintf("%t", threatData.IsAnonymous))
	displayTable(w, "Is Known Attacker", fmt.Sprintf("%t", threatData.IsKnownAttacker))
	displayTable(w, "Is Known Abuser", fmt.Sprintf("%t", threatData.IsKnownAbuser))
	displayTable(w, "Is Threat", fmt.Sprintf("%t", threatData.IsThreat))
	displayTable(w, "Is Bogon", fmt.Sprintf("%t", threatData.IsBogon))

	if len(threatData.Blocklists) > 0 {
		fmt.Fprintln(w, "Blocklists Found:")
		for i, blocklist := range threatData.Blocklists {
			displayTable(w, fmt.Sprintf("Blocklist %d Name", i+1), blocklist.Name)
			displayTable(w, "Site", blocklist.Site)
			displayTable(w, "Type", blocklist.Type)
			if i < len(threatData.Blocklists)-1 {
				fmt.Fprintln(w, "\t\t---\t")
			}
		}
	} else {
		displayTable(w, "Blocklists", "None")
	}

	w.Flush()
}

// Function to validate IP address format
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}
