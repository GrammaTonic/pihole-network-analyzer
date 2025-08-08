package pihole

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"pihole-analyzer/internal/types"
)

// QueryParams represents parameters for DNS query requests
type QueryParams struct {
	StartTime    time.Time
	EndTime      time.Time
	ClientFilter string
	DomainFilter string
	Limit        int
	StatusFilter []int
	TypeFilter   []int
}

// APIQueryRecord represents a DNS query record from Pi-hole API
type APIQueryRecord struct {
	Timestamp float64 `json:"timestamp"`
	Type      string  `json:"type"`
	Domain    string  `json:"domain"`
	Client    string  `json:"client"`
	Status    int     `json:"status"`
	Reply     float64 `json:"reply"`
	DNSSEC    string  `json:"dnssec"`
}

// QueriesResponse represents the API response for queries endpoint
type QueriesResponse struct {
	Data []APIQueryRecord `json:"data"`
	Took float64          `json:"took"`
}

// StatsResponse represents statistics from Pi-hole API
type StatsResponse struct {
	DomainsBlocked      int                `json:"domains_being_blocked"`
	DNSQueriesToday     int                `json:"dns_queries_today"`
	AdsBlockedToday     int                `json:"ads_blocked_today"`
	AdsPercentageToday  float64            `json:"ads_percentage_today"`
	UniqueClients       int                `json:"unique_clients"`
	QueriesForwarded    int                `json:"queries_forwarded"`
	QueriesCached       int                `json:"queries_cached"`
	ClientsOverTime     map[string]int     `json:"clients_over_time"`
	QueryTypesOverTime  map[string]int     `json:"query_types_over_time"`
	TopQueries          map[string]int     `json:"top_queries"`
	TopAds              map[string]int     `json:"top_ads"`
	TopClients          map[string]int     `json:"top_clients"`
	ForwardDestinations map[string]float64 `json:"forward_destinations"`
	QueryTypes          map[string]float64 `json:"querytypes"`
	ReplyTypes          map[string]float64 `json:"reply_NXDOMAIN"`
	Took                float64            `json:"took"`
}

// ClientInfo represents client information from Pi-hole API
type ClientInfo struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	MAC  string `json:"mac"`
}

// NetworkResponse represents network information response
type NetworkResponse struct {
	Network []ClientInfo `json:"network"`
	Took    float64      `json:"took"`
}

// GetDNSQueries retrieves DNS queries from Pi-hole API
// This replaces the SSH database query functionality
func (c *Client) GetDNSQueries(ctx context.Context, params QueryParams) ([]types.PiholeRecord, error) {
	c.Logger.Info("Fetching DNS queries from Pi-hole API: start_time=%s end_time=%s limit=%d",
		params.StartTime.Format(time.RFC3339), params.EndTime.Format(time.RFC3339), params.Limit)

	// Build query parameters
	queryParams := map[string]string{
		"from":  strconv.FormatInt(params.StartTime.Unix(), 10),
		"until": strconv.FormatInt(params.EndTime.Unix(), 10),
	}

	if params.ClientFilter != "" {
		queryParams["client"] = params.ClientFilter
	}

	if params.DomainFilter != "" {
		queryParams["domain"] = params.DomainFilter
	}

	if params.Limit > 0 {
		queryParams["limit"] = strconv.Itoa(params.Limit)
	}

	// Make API request
	resp, err := c.makeRequest(ctx, "GET", "/api/queries", queryParams, true)
	if err != nil {
		c.Logger.Error("Failed to fetch queries: %v", err)
		return nil, fmt.Errorf("failed to fetch queries: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.Logger.Error("API request failed: status_code=%d response=%s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// Parse response
	var queriesResp QueriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&queriesResp); err != nil {
		return nil, fmt.Errorf("failed to parse queries response: %w", err)
	}

	c.Logger.Info("Successfully fetched queries: query_count=%d took_seconds=%.2f",
		len(queriesResp.Data), queriesResp.Took)

	// Convert API records to our internal format
	records := make([]types.PiholeRecord, 0, len(queriesResp.Data))
	for _, apiRecord := range queriesResp.Data {
		record := types.PiholeRecord{
			Timestamp: strconv.FormatFloat(apiRecord.Timestamp, 'f', -1, 64),
			Client:    apiRecord.Client,
			Domain:    apiRecord.Domain,
			Status:    apiRecord.Status,
			// Note: HWAddr will be populated from network info separately
		}
		records = append(records, record)
	}

	return records, nil
}

// GetStatistics retrieves general statistics from Pi-hole API
func (c *Client) GetStatistics(ctx context.Context) (*StatsResponse, error) {
	c.Logger.Info("Fetching statistics from Pi-hole API")

	resp, err := c.makeRequest(ctx, "GET", "/api/stats", nil, true)
	if err != nil {
		c.Logger.Error("Failed to fetch statistics: %v", err)
		return nil, fmt.Errorf("failed to fetch statistics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.Logger.Error("Statistics API request failed: status_code=%d response=%s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("statistics API request failed with status %d", resp.StatusCode)
	}

	var stats StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to parse statistics response: %w", err)
	}

	c.Logger.Info("Successfully fetched statistics: dns_queries_today=%d ads_blocked_today=%d took=%.2f",
		stats.DNSQueriesToday, stats.AdsBlockedToday, stats.Took)

	return &stats, nil
}

// GetNetworkInfo retrieves network device information from Pi-hole API
// This replaces the SSH query for client information
func (c *Client) GetNetworkInfo(ctx context.Context) ([]ClientInfo, error) {
	c.Logger.Info("Fetching network information from Pi-hole API")

	resp, err := c.makeRequest(ctx, "GET", "/api/network", nil, true)
	if err != nil {
		c.Logger.Error("Failed to fetch network info: %v", err)
		return nil, fmt.Errorf("failed to fetch network info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.Logger.Error("Network API request failed: status_code=%d response=%s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("network API request failed with status %d", resp.StatusCode)
	}

	var networkResp NetworkResponse
	if err := json.NewDecoder(resp.Body).Decode(&networkResp); err != nil {
		return nil, fmt.Errorf("failed to parse network response: %w", err)
	}

	c.Logger.Info("Successfully fetched network info: client_count=%d took=%.2f",
		len(networkResp.Network), networkResp.Took)

	return networkResp.Network, nil
}

// GetTopClients retrieves top clients from Pi-hole API statistics
func (c *Client) GetTopClients(ctx context.Context) (map[string]int, error) {
	stats, err := c.GetStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get top clients: %w", err)
	}

	return stats.TopClients, nil
}

// GetTopDomains retrieves top domains from Pi-hole API statistics
func (c *Client) GetTopDomains(ctx context.Context) (map[string]int, error) {
	stats, err := c.GetStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get top domains: %w", err)
	}

	return stats.TopQueries, nil
}

// BuildClientStats builds client statistics from API data
// This matches the SSH implementation functionality
func (c *Client) BuildClientStats(ctx context.Context, queries []types.PiholeRecord, networkInfo []ClientInfo) (map[string]*types.ClientStats, error) {
	c.Logger.Info("Building client statistics from API data: query_count=%d network_count=%d",
		len(queries), len(networkInfo))

	clientStats := make(map[string]*types.ClientStats)

	// Create MAC address lookup map
	macLookup := make(map[string]string)
	for _, client := range networkInfo {
		if client.IP != "" && client.MAC != "" {
			macLookup[client.IP] = client.MAC
		}
	}

	// Process each query record
	for _, record := range queries {
		client := record.Client
		if client == "" {
			continue
		}

		// Initialize client stats if not exists
		if clientStats[client] == nil {
			hwAddr := macLookup[client] // Get MAC from network info
			clientStats[client] = &types.ClientStats{
				Client:      client,
				HWAddr:      hwAddr,
				Domains:     make(map[string]int),
				QueryTypes:  make(map[int]int),
				StatusCodes: make(map[int]int),
			}
		}

		stats := clientStats[client]
		stats.TotalQueries++

		// Track status codes
		stats.StatusCodes[record.Status]++

		// Track domain queries
		if record.Domain != "" {
			stats.Domains[record.Domain]++
			if stats.Domains[record.Domain] == 1 {
				stats.UniqueQueries++
			}
		}
	}

	c.Logger.Info("Client statistics built successfully: client_count=%d", len(clientStats))

	return clientStats, nil
}
