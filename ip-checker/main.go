package main

import (
	"context"
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
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

// Configuration holds all app settings
type Configuration struct {
	APIKey            string
	RequestTimeout    time.Duration
	ConnectTimeout    time.Duration
	MaxIdleConns      int
	IdleConnTimeout   time.Duration
	RequestsPerSecond float64
}

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
	IP               string
	ASNData          *ASNData
	ThreatData       *ThreatData
	BrowserLeaksData map[string]string
	Errors           map[string]error
	Success          bool
}

// Services holds all service dependencies
type Services struct {
	HTTPClient *http.Client
	Config     Configuration
	RateLimiter *rate.Limiter
}

func main() {
	ip := flag.String("ip", "", "IP address to lookup")
	format := flag.String("format", "table", "Output format: table, json")
	configFile := flag.String("config", "", "Path to config file (optional)")
	flag.Parse()

	if *ip == "" {
		logrus.Fatal("Please provide an IP address using the --ip flag")
	}

	if !isValidIP(*ip) {
		logrus.Fatalf("Invalid IP address: %s", *ip)
	}

	// Load configuration
	config := loadConfiguration(*configFile)
	
	// Initialize services
	services := initServices(config)

	// Fetch data
	results := fetchIPData(*ip, services)
	
	// Display results
	switch *format {
	case "json":
		displayJSONResults(results)
	case "table":
		displayTableResults(results)
	default:
		logrus.Fatalf("Unknown output format: %s", *format)
	}
}

func loadConfiguration(configFile string) Configuration {
	// Default configuration
	config := Configuration{
		APIKey:            os.Getenv("IPDATA_API_KEY"),
		RequestTimeout:    10 * time.Second,
		ConnectTimeout:    5 * time.Second,
		MaxIdleConns:      10,
		IdleConnTimeout:   90 * time.Second,
		RequestsPerSecond: 5.0,
	}

	// TODO: If configFile is provided, parse it and override defaults
	// For now, just validate we have the minimum requirements
	if config.APIKey == "" {
		logrus.Fatal("API key missing. Set IPDATA_API_KEY environment variable")
	}

	return config
}

func initServices(config Configuration) Services {
	// Create an HTTP client with connection pooling
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: config.ConnectTimeout,
		}).DialContext,
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConns,
		IdleConnTimeout:     config.IdleConnTimeout,
	}

	httpClient := &http.Client{
		Timeout:   config.RequestTimeout,
		Transport: transport,
	}

	// Create a rate limiter
	limiter := rate.NewLimiter(rate.Limit(config.RequestsPerSecond), 1)

	return Services{
		HTTPClient: httpClient,
		Config:     config,
		RateLimiter: limiter,
	}
}

// fetchIPData fetches all data concurrently with proper error handling
func fetchIPData(ip string, services Services) IPResults {
	results := IPResults{
		IP:     ip,
		Errors: make(map[string]error),
	}

	// Use errgroup for better error handling in goroutines
	g, ctx := errgroup.WithContext(context.Background())

	// Fetch ASN data
	g.Go(func() error {
		if err := services.RateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter error: %w", err)
		}
		
		url := fmt.Sprintf("https://api.ipdata.co/%s/asn?api-key=%s", ip, services.Config.APIKey)
		logrus.Info("Fetching ASN data...")
		
		var asnData ASNData
		err := fetchJSON(services.HTTPClient, url, &asnData)
		if err != nil {
			results.Errors["ASN"] = fmt.Errorf("failed to fetch ASN data: %w", err)
			return nil // Continue with other requests
		}
		
		results.ASNData = &asnData
		return nil
	})

	// Fetch Threat data
	g.Go(func() error {
		if err := services.RateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter error: %w", err)
		}
		
		url := fmt.Sprintf("https://api.ipdata.co/%s/threat?api-key=%s", ip, services.Config.APIKey)
		logrus.Info("Fetching Threat data...")
		
		var threatData ThreatData
		err := fetchJSON(services.HTTPClient, url, &threatData)
		if err != nil {
			results.Errors["Threat"] = fmt.Errorf("failed to fetch Threat data: %w", err)
			return nil // Continue with other requests
		}
		
		results.ThreatData = &threatData
		return nil
	})

	// Fetch BrowserLeaks data
	g.Go(func() error {
		if err := services.RateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter error: %w", err)
		}
		
		url := fmt.Sprintf("https://browserleaks.com/ip/%s", ip)
		logrus.Info("Fetching BrowserLeaks data...")
		
		data, err := fetchBrowserLeaksData(services.HTTPClient, url)
		if err != nil {
			results.Errors["BrowserLeaks"] = fmt.Errorf("failed to fetch BrowserLeaks data: %w", err)
			return nil // Continue with other requests
		}
		
		results.BrowserLeaksData = data
		return nil
	})

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		logrus.Errorf("Error during concurrent fetching: %v", err)
	}

	// Consider the operation successful if we have at least one data source
	results.Success = results.ASNData != nil || results.ThreatData != nil || len(results.BrowserLeaksData) > 0

	return results
}

// fetchJSON is a generic function to fetch and decode JSON
func fetchJSON[T any](client *http.Client, url string, result *T) error {
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		return fmt.Errorf("rate limit exceeded, retry after %s seconds", retryAfter)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("JSON decode failed: %w", err)
	}
	
	return nil
}

// fetchBrowserLeaksData scrapes data from browserleaks.com
func fetchBrowserLeaksData(client *http.Client, url string) (map[string]string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HTML parsing failed: %w", err)
	}

	fields := []string{"Country", "ISP", "Organization", "Usage Type"}
	data := make(map[string]string, len(fields)) // Pre-allocate with exact size
	
	for _, field := range fields {
		data[field] = doc.Find(fmt.Sprintf("td:contains('%s')", field)).Next().Text()
		if data[field] == "" {
			data[field] = "N/A"
		}
	}
	
	return data, nil
}

// displayTableResults shows all collected data in a tabulated format
func displayTableResults(results IPResults) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "\nIP Address: %s\n\n", results.IP)
	fmt.Fprintln(w, "Category\tValue")
	
	// Show errors if any
	if len(results.Errors) > 0 {
		fmt.Fprintln(w, "Errors:")
		for source, err := range results.Errors {
			fmt.Fprintf(w, "%s Error\t%v\n", source, err)
		}
		fmt.Fprintln(w, "---\t---")
	}

	// ASN Data
	if results.ASNData != nil {
		fmt.Fprintf(w, "ASN\t%s\n", results.ASNData.ASN)
		fmt.Fprintf(w, "ASN Name\t%s\n", results.ASNData.Name)
		fmt.Fprintf(w, "ASN Type\t%s\n", results.ASNData.Type)
	} else if _, ok := results.Errors["ASN"]; !ok {
		fmt.Fprintln(w, "ASN Data\tNot available")
	}

	// BrowserLeaks Data
	if len(results.BrowserLeaksData) > 0 {
		for key, value := range results.BrowserLeaksData {
			fmt.Fprintf(w, "%s\t%s\n", key, value)
		}
	} else if _, ok := results.Errors["BrowserLeaks"]; !ok {
		fmt.Fprintln(w, "BrowserLeaks Data\tNot available")
	}

	// Threat Data
	if results.ThreatData != nil {
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
	} else if _, ok := results.Errors["Threat"]; !ok {
		fmt.Fprintln(w, "Threat Data\tNot available")
	}

	w.Flush()
}

// displayJSONResults outputs the results in JSON format
func displayJSONResults(results IPResults) {
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		logrus.Errorf("Failed to marshal results to JSON: %v", err)
		return
	}
	fmt.Println(string(jsonData))
}

// isValidIP validates IP address format
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}
