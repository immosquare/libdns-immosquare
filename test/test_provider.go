package main

import (
	"context"
	"fmt"
	"log"
	"net/netip"
	"os"
	"time"
	
	"github.com/immosquare/libdns-immosquare"
	"github.com/libdns/libdns"
)

func main() {
	// Récupérer le token depuis la variable d'environnement
	apiToken := os.Getenv("IMMOSQUARE_API_TOKEN")
	if apiToken == "" {
		log.Fatal("Erreur: Variable d'environnement IMMOSQUARE_API_TOKEN non définie")
	}

	// Configuration du provider
	provider := &libdnsimmosquare.Provider{
		APIToken: apiToken,
		Endpoint: "https://immosquare.me:4005/api/dns",
	}

	ctx := context.Background()
	zone := "example.com" 

	fmt.Println("=== Test du provider DNS immosquare ===")

	// Test 1: Récupérer les enregistrements existants
	fmt.Println("\n1. Récupération des enregistrements existants...")
	records, err := provider.GetRecords(ctx, zone)
	if err != nil {
		log.Printf("Erreur GetRecords: %v", err)
	} else {
		fmt.Printf("✅ %d enregistrements trouvés\n", len(records))
		for i, record := range records {
			rr := record.RR()
			fmt.Printf("  %d. %s %s %s (TTL: %s)\n", i+1, rr.Name, rr.Type, rr.Data, rr.TTL)
		}
	}

	// Test 2: Ajouter un enregistrement TXT pour ACME challenge
	fmt.Println("\n2. Ajout d'un enregistrement TXT pour ACME challenge...")
	newTXTRecord := libdns.TXT{
		Name: "_acme-challenge",
		Text: "test-challenge-token-12345",
		TTL:  300 * time.Second,
	}

	addedRecords, err := provider.AppendRecords(ctx, zone, []libdns.Record{newTXTRecord})
	if err != nil {
		log.Printf("Erreur AppendRecords: %v", err)
	} else {
		fmt.Printf("✅ %d enregistrements ajoutés\n", len(addedRecords))
		for i, record := range addedRecords {
			rr := record.RR()
			fmt.Printf("  %d. %s %s %s (TTL: %s)\n", 
				i+1, rr.Name, rr.Type, rr.Data, rr.TTL)
		}
	}

	// Test 3: Utiliser SetRecords pour définir tous les enregistrements
	fmt.Println("\n3. Test SetRecords (remplacer tous les enregistrements)...")
	
	// Créer des adresses IP valides
	ip1, _ := netip.ParseAddr("192.99.250.180")
	ip2, _ := netip.ParseAddr("192.99.250.181")
	ip3, _ := netip.ParseAddr("192.99.250.182")
	
	setRecords := []libdns.Record{
		libdns.Address{
			Name: "www",
			IP:   ip1,
			TTL:  600 * time.Second,
		},
		libdns.Address{
			Name: "mail",
			IP:   ip2,
			TTL:  900 * time.Second,
		},
		libdns.Address{
			Name: "api",
			IP:   ip3,
			TTL:  1200 * time.Second,
		},
	}

	updatedRecords, err := provider.SetRecords(ctx, zone, setRecords)
	if err != nil {
		log.Printf("Erreur SetRecords: %v", err)
	} else {
		fmt.Printf("✅ %d enregistrements définis\n", len(updatedRecords))
		for i, record := range updatedRecords {
			rr := record.RR()
			fmt.Printf("  %d. %s %s %s (TTL: %s)\n", 
				i+1, rr.Name, rr.Type, rr.Data, rr.TTL)
		}
	}

	// Test 4: Supprimer un enregistrement
	fmt.Println("\n4. Test DeleteRecords (supprimer un enregistrement)...")
	deleteRecords := []libdns.Record{
		libdns.Address{
			Name: "api",
			IP:   ip3,
			TTL:  1200 * time.Second,
		},
	}

	deletedRecords, err := provider.DeleteRecords(ctx, zone, deleteRecords)
	if err != nil {
		log.Printf("Erreur DeleteRecords: %v", err)
	} else {
		fmt.Printf("✅ %d enregistrements supprimés\n", len(deletedRecords))
	}

	// Test 5: Test avec différents types d'enregistrements
	fmt.Println("\n5. Test avec différents types d'enregistrements...")
	
	// Enregistrement CNAME
	cnameRecord := libdns.CNAME{
		Name:   "www2",
		Target: "www.example.com",
		TTL:    300 * time.Second,
	}
	
	// Enregistrement MX
	mxRecord := libdns.MX{
		Name:       "@",
		Preference: 10,
		Target:     "mail.example.com",
		TTL:        600 * time.Second,
	}
	
	// Enregistrement NS
	nsRecord := libdns.NS{
		Name:   "@",
		Target: "ns1.example.com",
		TTL:    86400 * time.Second,
	}
	
	mixedRecords := []libdns.Record{cnameRecord, mxRecord, nsRecord}
	
	addedMixedRecords, err := provider.AppendRecords(ctx, zone, mixedRecords)
	if err != nil {
		log.Printf("Erreur AppendRecords (types mixtes): %v", err)
	} else {
		fmt.Printf("✅ %d enregistrements mixtes ajoutés\n", len(addedMixedRecords))
		for i, record := range addedMixedRecords {
			rr := record.RR()
			fmt.Printf("  %d. %s %s %s (TTL: %s)\n", 
				i+1, rr.Name, rr.Type, rr.Data, rr.TTL)
		}
	}

	fmt.Println("\n=== Test terminé ===")
} 