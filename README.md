# Immosquare DNS pour [`libdns`](https://github.com/libdns/libdns)

[![Go Reference](https://pkg.go.dev/badge/github.com/immosquare/libdns-immosquare.svg)](https://pkg.go.dev/github.com/immosquare/libdns-immosquare)

Ce package implémente les [interfaces libdns](https://github.com/libdns/libdns) pour le service DNS immosquare, vous permettant de gérer les enregistrements DNS pour les validations ACME de Caddy.

## Configuration

### Paramètres du Provider

- `api_token` (string, requis) : Token d'authentification pour l'API immosquare
- `endpoint` (string, optionnel) : Endpoint de l'API DNS (par défaut: `https://immosquare.me:4005/api/dns`)

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

## Types d'enregistrements supportés

- A
- AAAA
- CNAME
- MX
- TXT
- NS
- SOA

## Utilisation pour les validations ACME

Ce provider est particulièrement adapté pour les validations ACME de Caddy, permettant la génération automatique de certificats SSL/TLS via des enregistrements TXT `_acme-challenge`.

## Licence

Ce projet est sous licence Apache 2.0.
