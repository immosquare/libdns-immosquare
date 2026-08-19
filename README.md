---
locale: en
tags:
  - app:libdns-immosquare
  - audience:technique
---

# libdns-immosquare

`libdns-immosquare` is a generic DNS provider for [`libdns`](https://github.com/libdns/libdns) that works with any compatible DNS API. This page covers installing the Go package, configuring the provider, the HTTP endpoints and record types the DNS API has to support, and the minimum TTL applied when records are written. The package reference lives on [`Go Package`](https://pkg.go.dev/github.com/immosquare/libdns-immosquare), and the code is released under the MIT license.

## Installing, configuring and testing the libdns-immosquare provider

Install the package with `go get`:

```bash
go get github.com/immosquare/libdns-immosquare
```

Build a `Provider` with the base URL of your DNS API and, when that API requires one, an API token:

```go
provider := &libdnsimmosquare.Provider{
    APIToken: "your-api-token",
    Endpoint: "https://your-dns-api.com/api/dns",
}
```

The `Provider` struct carries two fields:

| Field      | Type     | Required | Description                                 |
| ---------- | -------- | -------- | ------------------------------------------- |
| `Endpoint` | `string` | yes      | Base URL of the DNS API (no trailing slash) |
| `APIToken` | `string` | no       | Sent as `Authorization: Bearer <token>`     |

Exercise the provider against a live DNS API with the bundled test script:

```bash
API_TOKEN=your-api-token ENDPOINT=https://your-dns-api.com/api/dns go run test/test_provider.go
```

## Endpoints and record types the DNS API must support for libdns-immosquare

Your DNS API must expose these endpoints:

```
GET    /zones/{domain}/records
POST   /zones/{domain}/records
PUT    /zones/{domain}/records
DELETE /zones/{domain}/records
```

`libdns-immosquare` supports the following record types:

- **A/AAAA** : `libdns.Address` with `IP` field of type `netip.Addr`
- **TXT** : `libdns.TXT` with `Text` field
- **CNAME** : `libdns.CNAME` with `Target` field
- **MX** : `libdns.MX` with `Preference` and `Target` fields
- **NS** : `libdns.NS` with `Target` field
- **Other types** : `libdns.RR` for unsupported record types

## Minimum TTL of 120 seconds in libdns-immosquare

`AppendRecords` and `SetRecords` clamp any TTL below 120 seconds up to 120 seconds. This prevents records created with `TTL: 0` (e.g. certmagic ACME challenges) from inheriting a high zone default like 1800s and slowing down DNS propagation. `DeleteRecords` does not apply the clamp.
