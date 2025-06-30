package main

import (
	"fmt"
	"github.com/immosquare/libdns-immosquare"
)

func main() {
	fmt.Printf("Version du provider libdns-immosquare: %s\n", libdnsimmosquare.Version)
	
	// Exemple d'utilisation du provider avec la version
	provider := &libdnsimmosquare.Provider{
		APIToken: "your-token-here",
		Endpoint: "https://monitoring.immosquare.com/api/dns",
	}
	
	fmt.Printf("Provider configuré avec la version %s\n", libdnsimmosquare.Version)
	_ = provider // Éviter l'erreur de variable non utilisée
} 