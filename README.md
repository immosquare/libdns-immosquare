# libdns-immosquare

[`Go Package`](https://pkg.go.dev/github.com/immosquare/libdns-immosquare)

A generic DNS provider for [`libdns`](https://github.com/libdns/libdns) that works with any compatible DNS API.

## Installation

```bash
go get github.com/immosquare/libdns-immosquare
```

## Configuration

```go
provider := &immosquare.Provider{
    APIToken: "your-api-token",
    Endpoint: "https://your-dns-api.com/api/dns",
}
```

## Required API Endpoints

Your DNS API must expose these endpoints:

```
GET    /zones/{domain}/records
POST   /zones/{domain}/records  
PUT    /zones/{domain}/records
DELETE /zones/{domain}/records
```

## Supported Record Types

- **A/AAAA** : `libdns.Address` with `IP` field of type `netip.Addr`
- **TXT** : `libdns.TXT` with `Text` field
- **CNAME** : `libdns.CNAME` with `Target` field
- **MX** : `libdns.MX` with `Preference` and `Target` fields
- **NS** : `libdns.NS` with `Target` field
- **Other types** : `libdns.RR` for unsupported record types

## Test

```bash
API_TOKEN=your-api-token ENDPOINT=https://your-dns-api.com/api/dns go run test/test_provider.go
```

## License

MIT
