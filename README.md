# Immosquare DNS pour [`libdns`](https://github.com/libdns/libdns)

[![Go Reference](https://pkg.go.dev/badge/github.com/immosquare/libdns-immosquare.svg)](https://pkg.go.dev/github.com/immosquare/libdns-immosquare)

Ce package implémente les [interfaces libdns](https://github.com/libdns/libdns) pour le service DNS immosquare, vous permettant de gérer les enregistrements DNS pour les validations ACME de Caddy.

**✅ Compatible avec libdns v1.0.0**

## Configuration

### Paramètres du Provider

- `api_token` (string, requis) : Token d'authentification pour l'API immosquare
- `endpoint` (string, optionnel) : Endpoint de l'API DNS (par défaut: `https://monitoring.immosquare.com/api/dns`)

### Exemple de configuration Caddy

```json
{
  "dns": {
    "provider": "immosquare",
    "api_token": "votre_token_ici",
    "endpoint": "https://immosquare.me:4005/api/dns"
  }
}
```

### Exemple d'utilisation avec Caddy

```caddyfile
example.com {
    tls {
        dns immosquare {
            api_token "votre_token_ici"
        }
    }
}
```

## Fonctionnalités

- ✅ Récupération d'enregistrements DNS (`GetRecords`)
- ✅ Ajout d'enregistrements DNS (`AppendRecords`)
- ✅ Mise à jour d'enregistrements DNS (`SetRecords`)
- ✅ Suppression d'enregistrements DNS (`DeleteRecords`)
- ✅ Support des validations ACME pour Caddy
- ✅ Gestion automatique des timeouts et erreurs
- ✅ Support complet de libdns v1.0.0 avec les nouvelles structures de données

## Types d'enregistrements supportés

### Structures libdns v1.0.0 utilisées

- **A/AAAA** : `libdns.Address` avec champ `IP` de type `netip.Addr`
- **TXT** : `libdns.TXT` avec champ `Text`
- **CNAME** : `libdns.CNAME` avec champ `Target`
- **MX** : `libdns.MX` avec champs `Preference` et `Target`
- **NS** : `libdns.NS` avec champ `Target`
- **Autres types** : `libdns.RR` pour les types non spécifiquement supportés

### Exemples d'utilisation

```go
// Enregistrement A
address := libdns.Address{
    Name: "www",
    IP:   netip.MustParseAddr("192.168.1.1"),
    TTL:  300 * time.Second,
}

// Enregistrement TXT
txt := libdns.TXT{
    Name: "_acme-challenge",
    Text: "challenge-token",
    TTL:  300 * time.Second,
}

// Enregistrement MX
mx := libdns.MX{
    Name:       "@",
    Preference: 10,
    Target:     "mail.example.com",
    TTL:        600 * time.Second,
}
```

## Utilisation pour les validations ACME

Ce provider est particulièrement adapté pour les validations ACME de Caddy, permettant la génération automatique de certificats SSL/TLS via des enregistrements TXT `_acme-challenge`.

## Licence

Ce projet est sous licence MIT.
