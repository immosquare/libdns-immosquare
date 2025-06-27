package main

import (
	"context"
	"fmt"
	"log"
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
	newTXTRecord := libdns.RR{
		Name: "_acme-challenge",
		Type: "TXT",
		Data: "test-challenge-token-12345",
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
	setRecords := []libdns.Record{
		libdns.RR{
			Name: "www",
			Type: "A",
			Data: "192.99.250.180",
			TTL:  600 * time.Second,
		},
		libdns.RR{
			Name: "mail",
			Type: "A", 
			Data: "192.99.250.181",
			TTL:  900 * time.Second,
		},
		libdns.RR{
			Name: "api",
			Type: "A",
			Data: "192.99.250.182", 
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
		libdns.RR{
			Name: "api",
			Type: "A",
			Data: "192.99.250.182", 
			TTL:  1200 * time.Second,
		},
	}

	deletedRecords, err := provider.DeleteRecords(ctx, zone, deleteRecords)
	if err != nil {
		log.Printf("Erreur DeleteRecords: %v", err)
	} else {
		fmt.Printf("✅ %d enregistrements supprimés\n", len(deletedRecords))
	}

	fmt.Println("\n=== Test terminé ===")
} 