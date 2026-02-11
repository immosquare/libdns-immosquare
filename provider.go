// Package libdnsimmosquare implements a DNS records management client
package libdnsimmosquare

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/libdns/libdns"
)

// Version of the libdns-immosquare provider
const Version = "1.0.4"

// defaultMinTTL is the minimum TTL applied to records created via this provider.
// Prevents issues with TTL 0 (e.g. certmagic ACME challenges) falling back to
// high zone defaults like 1800s, which slows down DNS propagation.
const defaultMinTTL = 120 * time.Second


type Provider struct {
	APIToken string `json:"api_token,omitempty"`
	Endpoint string `json:"endpoint"`
	client *http.Client
}

// initClient initializes the HTTP client if necessary
func (p *Provider) initClient() error {
	if p.client == nil {
		p.client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	if p.Endpoint == "" {
		return fmt.Errorf("endpoint is required for the immosquare provider")
	}
	return nil
}

// makeRequest makes an HTTP request to the immosquare API
func (p *Provider) makeRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	if err := p.initClient(); err != nil {
		return nil, err
	}
	
	url := p.Endpoint + path
	var req *http.Request
	var err error
	
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("JSON serialization error: %w", err)
		}
		req, err = http.NewRequestWithContext(ctx, method, url, strings.NewReader(string(jsonBody)))
		if err != nil {
			return nil, fmt.Errorf("request creation error: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return nil, fmt.Errorf("request creation error: %w", err)
		}
	}
	
	// Add authentication token
	if p.APIToken != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIToken)
	}
	
	return p.client.Do(req)
}

// GetRecords retrieves all DNS records for the specified zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	resp, err := p.makeRequest(ctx, "GET", "/zones/"+zone+"/records", nil)
	if err != nil {
		return nil, fmt.Errorf("GET request error: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}
	
	// Read the raw response to see the structure
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("body reading error: %w", err)
	}
	
	// Try to decode as an object with a records field
	var apiResponse struct {
		Records []struct {
			Name  string `json:"name"`
			Type  string `json:"type"`
			Value string `json:"value"`
			TTL   int    `json:"ttl"`
		} `json:"records"`
	}
	
	if err := json.Unmarshal(bodyBytes, &apiResponse); err != nil {
		// If it doesn't work, try as a direct array
		var apiRecords []struct {
			Name  string `json:"name"`
			Type  string `json:"type"`
			Value string `json:"value"`
			TTL   int    `json:"ttl"`
		}
		
		if err := json.Unmarshal(bodyBytes, &apiRecords); err != nil {
			return nil, fmt.Errorf("JSON decoding error: %w", err)
		}
		
		records := make([]libdns.Record, 0, len(apiRecords))
		for _, apiRecord := range apiRecords {
			record, err := p.convertAPIRecordToLibDNS(apiRecord)
			if err != nil {
				return nil, fmt.Errorf("record conversion error: %w", err)
			}
			records = append(records, record)
		}
		return records, nil
	}
	
	// Utiliser la réponse avec le champ records
	records := make([]libdns.Record, 0, len(apiResponse.Records))
	for _, apiRecord := range apiResponse.Records {
		record, err := p.convertAPIRecordToLibDNS(apiRecord)
		if err != nil {
			return nil, fmt.Errorf("record conversion error: %w", err)
		}
		records = append(records, record)
	}
	
	return records, nil
}

// convertAPIRecordToLibDNS converts an API record to the appropriate libdns structure
func (p *Provider) convertAPIRecordToLibDNS(apiRecord struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}) (libdns.Record, error) {
	ttl := time.Duration(apiRecord.TTL) * time.Second
	
	switch strings.ToUpper(apiRecord.Type) {
	case "A", "AAAA":
		ip, err := netip.ParseAddr(apiRecord.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid IP address '%s': %w", apiRecord.Value, err)
		}
		address := libdns.Address{
			Name: apiRecord.Name,
			TTL:  ttl,
			IP:   ip,
		}
		return address, nil
	case "TXT":
		txt := libdns.TXT{
			Name: apiRecord.Name,
			Text: apiRecord.Value,
			TTL:  ttl,
		}
		return txt, nil
	case "CNAME":
		cname := libdns.CNAME{
			Name:   apiRecord.Name,
			Target: apiRecord.Value,
			TTL:    ttl,
		}
		return cname, nil
	case "MX":
		// For MX records, we need to parse the priority and target
		// Expected format: "10 mail.example.com" or just "mail.example.com"
		parts := strings.Fields(apiRecord.Value)
		var preference uint16 = 10
		var target string
		
		if len(parts) >= 2 {
			// Format: "10 mail.example.com"
			if pref, err := parseUint16(parts[0]); err == nil {
				preference = pref
				target = strings.Join(parts[1:], " ")
			} else {
				// Format: "mail.example.com" (no priority)
				target = apiRecord.Value
			}
		} else {
			// Format: "mail.example.com"
			target = apiRecord.Value
		}
		
		mx := libdns.MX{
			Name:       apiRecord.Name,
			Preference: preference,
			Target:     target,
			TTL:        ttl,
		}
		return mx, nil
	case "NS":
		ns := libdns.NS{
			Name:   apiRecord.Name,
			Target: apiRecord.Value,
			TTL:    ttl,
		}
		return ns, nil
	default:
		rr := libdns.RR{
			Name: apiRecord.Name,
			Type: apiRecord.Type,
			Data: apiRecord.Value,
			TTL:  ttl,
		}
		return rr, nil
	}
}

// parseUint16 parses a string to uint16
func parseUint16(s string) (uint16, error) {
	var result uint16
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// convertToSpecificTypes converts records to specific types
func (p *Provider) convertToSpecificTypes(records []libdns.Record) []libdns.Record {
	result := make([]libdns.Record, 0, len(records))
	for _, record := range records {
		rr := record.RR()
		switch strings.ToUpper(rr.Type) {
		case "A", "AAAA":
			ip, err := netip.ParseAddr(rr.Data)
			if err != nil {
				// If the IP is not valid, keep the RR
				result = append(result, rr)
				continue
			}
			address := libdns.Address{
				Name: rr.Name,
				TTL:  rr.TTL,
				IP:   ip,
			}
			result = append(result, address)
		case "TXT":
			txt := libdns.TXT{
				Name: rr.Name,
				Text: rr.Data,
				TTL:  rr.TTL,
			}
			result = append(result, txt)
		case "CNAME":
			cname := libdns.CNAME{
				Name:   rr.Name,
				Target: rr.Data,
				TTL:    rr.TTL,
			}
			result = append(result, cname)
		case "MX":
			// Parse the priority and target for MX
			parts := strings.Fields(rr.Data)
			var preference uint16 = 10
			var target string
			
			if len(parts) >= 2 {
				if pref, err := parseUint16(parts[0]); err == nil {
					preference = pref
					target = strings.Join(parts[1:], " ")
				} else {
					target = rr.Data
				}
			} else {
				target = rr.Data
			}
			
			mx := libdns.MX{
				Name:       rr.Name,
				Preference: preference,
				Target:     target,
				TTL:        rr.TTL,
			}
			result = append(result, mx)
		case "NS":
			ns := libdns.NS{
				Name:   rr.Name,
				Target: rr.Data,
				TTL:    rr.TTL,
			}
			result = append(result, ns)
		default:
			result = append(result, rr)
		}
	}
	return result
}

// AppendRecords adds new DNS records to the zone.
// Returns the records that have been added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	if len(records) == 0 {
		return []libdns.Record{}, nil
	}
	
	// Convert records to API format according to the type
	apiRecords := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		rr := record.RR()
		ttl := rr.TTL
		if ttl < defaultMinTTL {
			ttl = defaultMinTTL
		}
		apiRecord := map[string]interface{}{
			"name": rr.Name,
			"type": rr.Type,
			"data": rr.Data, // The API expects "data" for all types
			"ttl":  int(ttl.Seconds()),
		}

		apiRecords = append(apiRecords, apiRecord)
	}

	// Send as an object with a records field
	requestBody := map[string]interface{}{
		"records": apiRecords,
	}

	resp, err := p.makeRequest(ctx, "POST", "/zones/"+zone+"/records", requestBody)
	if err != nil {
		return nil, fmt.Errorf("POST request error: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error during addition: %s", resp.Status)
	}
	
	// Return the records converted to specific types
	return p.convertToSpecificTypes(records), nil
}

// SetRecords sets the DNS records in the zone, updating existing records or creating new ones.
// Returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	if len(records) == 0 {
		return []libdns.Record{}, nil
	}
	
	// Convert records to API format according to the type
	apiRecords := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		rr := record.RR()
		ttl := rr.TTL
		if ttl < defaultMinTTL {
			ttl = defaultMinTTL
		}
		apiRecord := map[string]interface{}{
			"name": rr.Name,
			"type": rr.Type,
			"data": rr.Data, // The API expects "data" for all types
			"ttl":  int(ttl.Seconds()),
		}

		apiRecords = append(apiRecords, apiRecord)
	}

	// Send as an object with a records field
	requestBody := map[string]interface{}{
		"records": apiRecords,
	}

	resp, err := p.makeRequest(ctx, "PUT", "/zones/"+zone+"/records", requestBody)
	if err != nil {
		return nil, fmt.Errorf("PUT request error: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error during update: %s", resp.Status)
	}
	
	// Return the records converted to specific types
	return p.convertToSpecificTypes(records), nil
}

// DeleteRecords deletes the specified DNS records from the zone.
// Returns the records that have been deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	if len(records) == 0 {
		return []libdns.Record{}, nil
	}
	
	// Convert records to API format according to the type
	apiRecords := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		rr := record.RR()
		apiRecord := map[string]interface{}{
			"name": rr.Name,
			"type": rr.Type,
			"data": rr.Data, // The API expects "data" for all types
			"ttl":  int(rr.TTL.Seconds()),
		}
		
		apiRecords = append(apiRecords, apiRecord)
	}
	
	// Envoyer les enregistrements à supprimer dans le body
	requestBody := map[string]interface{}{
		"records": apiRecords,
	}
	
	resp, err := p.makeRequest(ctx, "DELETE", "/zones/"+zone+"/records", requestBody)
	if err != nil {
		return nil, fmt.Errorf("DELETE request error: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		// Return the records converted to specific types
		return p.convertToSpecificTypes(records), nil
	}
	
	return []libdns.Record{}, nil
}

// Interface guards to ensure the Provider implements all libdns interfaces
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
