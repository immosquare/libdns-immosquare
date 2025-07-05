# Provider DNS générique pour [`libdns`](https://github.com/libdns/libdns)

[![Go Reference](https://pkg.go.dev/badge/github.com/immosquare/libdns-immosquare.svg)](https://pkg.go.dev/github.com/immosquare/libdns-immosquare)

Ce package implémente les [interfaces libdns](https://github.com/libdns/libdns) pour n'importe quelle API DNS compatible, vous permettant de gérer les enregistrements DNS pour les validations ACME de Caddy.

**✅ Compatible avec libdns v1.x**


## Installation

```bash
go get github.com/immosquare/libdns-immosquare
```

## Configuration

### Paramètres du Provider

- `api_token` (string, requis) : Token d'authentification pour l'API DNS
- `endpoint` (string, requis) : Endpoint de l'API DNS (ex: `https://votre-api.com/api/dns`)

### Format d'API requis

Ce provider est compatible avec n'importe quelle API DNS qui expose les endpoints suivants :

```
GET    /zones/{domain}/records    - Récupérer tous les enregistrements d'une zone
POST   /zones/{domain}/records    - Créer de nouveaux enregistrements
PUT    /zones/{domain}/records    - Mettre à jour des enregistrements existants
DELETE /zones/{domain}/records    - Supprimer des enregistrements
```

Où `{domain}` est le nom de domaine de la zone DNS.


## Fonctionnalités

- ✅ Récupération d'enregistrements DNS (`GetRecords`)
- ✅ Ajout d'enregistrements DNS (`AppendRecords`)
- ✅ Mise à jour d'enregistrements DNS (`SetRecords`)
- ✅ Suppression d'enregistrements DNS (`DeleteRecords`)
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
