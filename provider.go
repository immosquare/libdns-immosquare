// Package libdnsimmosquare implémente un client de gestion d'enregistrements DNS
// compatible avec les interfaces libdns pour le service DNS immosquare.
// Ce package permet de gérer les enregistrements DNS via l'API immosquare
// pour les validations ACME de Caddy.
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

// Version du provider libdns-immosquare
const Version = "1.0.2"

// TODO: Providers must not require additional provisioning steps by the callers; it
// should work simply by populating a struct and calling methods on it. If your DNS
// service requires long-lived state or some extra provisioning step, do it implicitly
// when methods are called; sync.Once can help with this, and/or you can use a
// sync.(RW)Mutex in your Provider struct to synchronize implicit provisioning.

// Provider facilite la manipulation d'enregistrements DNS avec immosquare.
// Il utilise l'API REST pour gérer les enregistrements DNS.
type Provider struct {
	// Token d'authentification pour l'API immosquare
	APIToken string `json:"api_token,omitempty"`
	// Endpoint de l'API DNS (requis)
	Endpoint string `json:"endpoint"`
	// Client HTTP pour les requêtes API
	client *http.Client
}

// initClient initialise le client HTTP si nécessaire
func (p *Provider) initClient() error {
	if p.client == nil {
		p.client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	if p.Endpoint == "" {
		return fmt.Errorf("endpoint est requis pour le provider immosquare")
	}
	return nil
}

// makeRequest effectue une requête HTTP vers l'API immosquare
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
			return nil, fmt.Errorf("erreur de sérialisation JSON: %w", err)
		}
		req, err = http.NewRequestWithContext(ctx, method, url, strings.NewReader(string(jsonBody)))
		if err != nil {
			return nil, fmt.Errorf("erreur de création de requête: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return nil, fmt.Errorf("erreur de création de requête: %w", err)
		}
	}
	
	// Ajout du token d'authentification
	if p.APIToken != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIToken)
	}
	
	return p.client.Do(req)
}

// GetRecords récupère tous les enregistrements DNS de la zone spécifiée.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	resp, err := p.makeRequest(ctx, "GET", "/zones/"+zone+"/records", nil)
	if err != nil {
		return nil, fmt.Errorf("erreur de requête GET: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur API: %s", resp.Status)
	}
	
	// Lire la réponse brute pour voir la structure
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erreur de lecture du body: %w", err)
	}
	
	// Essayer de décoder comme un objet avec un champ records
	var apiResponse struct {
		Records []struct {
			Name  string `json:"name"`
			Type  string `json:"type"`
			Value string `json:"value"`
			TTL   int    `json:"ttl"`
		} `json:"records"`
	}
	
	if err := json.Unmarshal(bodyBytes, &apiResponse); err != nil {
		// Si ça ne marche pas, essayer comme un tableau direct
		var apiRecords []struct {
			Name  string `json:"name"`
			Type  string `json:"type"`
			Value string `json:"value"`
			TTL   int    `json:"ttl"`
		}
		
		if err := json.Unmarshal(bodyBytes, &apiRecords); err != nil {
			return nil, fmt.Errorf("erreur de décodage JSON: %w", err)
		}
		
		records := make([]libdns.Record, 0, len(apiRecords))
		for _, apiRecord := range apiRecords {
			record, err := p.convertAPIRecordToLibDNS(apiRecord)
			if err != nil {
				return nil, fmt.Errorf("erreur de conversion d'enregistrement: %w", err)
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
			return nil, fmt.Errorf("erreur de conversion d'enregistrement: %w", err)
		}
		records = append(records, record)
	}
	
	return records, nil
}

// convertAPIRecordToLibDNS convertit un enregistrement API en structure libdns appropriée
func (p *Provider) convertAPIRecordToLibDNS(apiRecord struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}) (libdns.Record, error) {
	ttl := time.Duration(apiRecord.TTL) * time.Second
	
	switch strings.ToUpper(apiRecord.Type) {
	case "A", "AAAA":
		// Pour les enregistrements A/AAAA, nous utilisons le type Address
		ip, err := netip.ParseAddr(apiRecord.Value)
		if err != nil {
			return nil, fmt.Errorf("adresse IP invalide '%s': %w", apiRecord.Value, err)
		}
		address := libdns.Address{
			Name: apiRecord.Name,
			TTL:  ttl,
			IP:   ip,
		}
		return address, nil
	case "TXT":
		// Pour les enregistrements TXT, nous utilisons le type TXT
		txt := libdns.TXT{
			Name: apiRecord.Name,
			Text: apiRecord.Value,
			TTL:  ttl,
		}
		return txt, nil
	case "CNAME":
		// Pour les enregistrements CNAME
		cname := libdns.CNAME{
			Name:   apiRecord.Name,
			Target: apiRecord.Value,
			TTL:    ttl,
		}
		return cname, nil
	case "MX":
		// Pour les enregistrements MX, nous devons parser la priorité et la cible
		// Format attendu: "10 mail.example.com" ou juste "mail.example.com"
		parts := strings.Fields(apiRecord.Value)
		var preference uint16 = 10 // Valeur par défaut
		var target string
		
		if len(parts) >= 2 {
			// Format: "10 mail.example.com"
			if pref, err := parseUint16(parts[0]); err == nil {
				preference = pref
				target = strings.Join(parts[1:], " ")
			} else {
				// Format: "mail.example.com" (pas de priorité)
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
		// Pour les enregistrements NS
		ns := libdns.NS{
			Name:   apiRecord.Name,
			Target: apiRecord.Value,
			TTL:    ttl,
		}
		return ns, nil
	default:
		// Pour les autres types, nous utilisons RR
		rr := libdns.RR{
			Name: apiRecord.Name,
			Type: apiRecord.Type,
			Data: apiRecord.Value,
			TTL:  ttl,
		}
		return rr, nil
	}
}

// parseUint16 parse une chaîne en uint16
func parseUint16(s string) (uint16, error) {
	var result uint16
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// convertToSpecificTypes convertit les enregistrements en types spécifiques
func (p *Provider) convertToSpecificTypes(records []libdns.Record) []libdns.Record {
	result := make([]libdns.Record, 0, len(records))
	for _, record := range records {
		rr := record.RR()
		switch strings.ToUpper(rr.Type) {
		case "A", "AAAA":
			// Pour les enregistrements A/AAAA, nous utilisons le type Address
			ip, err := netip.ParseAddr(rr.Data)
			if err != nil {
				// Si l'IP n'est pas valide, on garde RR
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
			// Parser la priorité et la cible pour MX
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

// AppendRecords ajoute de nouveaux enregistrements DNS à la zone.
// Retourne les enregistrements qui ont été ajoutés.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	if len(records) == 0 {
		return []libdns.Record{}, nil
	}
	
	// Conversion des enregistrements en format API selon le type
	apiRecords := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		rr := record.RR()
		apiRecord := map[string]interface{}{
			"name": rr.Name,
			"type": rr.Type,
			"data": rr.Data, // L'API attend "data" pour tous les types
			"ttl":  int(rr.TTL.Seconds()),
		}
		
		apiRecords = append(apiRecords, apiRecord)
	}
	
	// Envoyer comme un objet avec un champ records
	requestBody := map[string]interface{}{
		"records": apiRecords,
	}
	
	resp, err := p.makeRequest(ctx, "POST", "/zones/"+zone+"/records", requestBody)
	if err != nil {
		return nil, fmt.Errorf("erreur de requête POST: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur API lors de l'ajout: %s", resp.Status)
	}
	
	// Retourner les enregistrements convertis en types spécifiques
	return p.convertToSpecificTypes(records), nil
}

// SetRecords définit les enregistrements DNS dans la zone, en mettant à jour
// les enregistrements existants ou en créant de nouveaux.
// Retourne les enregistrements mis à jour.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	if len(records) == 0 {
		return []libdns.Record{}, nil
	}
	
	// Conversion des enregistrements en format API selon le type
	apiRecords := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		rr := record.RR()
		apiRecord := map[string]interface{}{
			"name": rr.Name,
			"type": rr.Type,
			"data": rr.Data, // L'API attend "data" pour tous les types
			"ttl":  int(rr.TTL.Seconds()),
		}
		
		apiRecords = append(apiRecords, apiRecord)
	}
	
	// Envoyer comme un objet avec un champ records
	requestBody := map[string]interface{}{
		"records": apiRecords,
	}
	
	resp, err := p.makeRequest(ctx, "PUT", "/zones/"+zone+"/records", requestBody)
	if err != nil {
		return nil, fmt.Errorf("erreur de requête PUT: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur API lors de la mise à jour: %s", resp.Status)
	}
	
	// Retourner les enregistrements convertis en types spécifiques
	return p.convertToSpecificTypes(records), nil
}

// DeleteRecords supprime les enregistrements DNS spécifiés de la zone.
// Retourne les enregistrements qui ont été supprimés.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	if len(records) == 0 {
		return []libdns.Record{}, nil
	}
	
	// Conversion des enregistrements en format API selon le type
	apiRecords := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		rr := record.RR()
		apiRecord := map[string]interface{}{
			"name": rr.Name,
			"type": rr.Type,
			"data": rr.Data, // L'API attend "data" pour tous les types
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
		return nil, fmt.Errorf("erreur de requête DELETE: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		// Retourner les enregistrements convertis en types spécifiques
		return p.convertToSpecificTypes(records), nil
	}
	
	return []libdns.Record{}, nil
}

// Interface guards pour s'assurer que le Provider implémente toutes les interfaces libdns
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
