# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A libdns provider implementation that enables DNS record management through a configurable HTTP API endpoint. This package implements the standard libdns interfaces for use with tools like Caddy's automatic HTTPS.

## Commands

**Run tests:**
```bash
API_TOKEN=your-token ENDPOINT=https://your-dns-api.com/api/dns go run test/test_provider.go
```

**Build/verify compilation:**
```bash
go build ./...
```

## Architecture

The provider (`provider.go`) implements four libdns interfaces:
- `RecordGetter` - GET /zones/{domain}/records
- `RecordAppender` - POST /zones/{domain}/records
- `RecordSetter` - PUT /zones/{domain}/records
- `RecordDeleter` - DELETE /zones/{domain}/records

**Record Type Handling:**
- API responses are converted to typed libdns structs (`libdns.Address`, `libdns.TXT`, `libdns.CNAME`, `libdns.MX`, `libdns.NS`)
- Outgoing records are normalized via `.RR()` to generic format before API calls
- Unsupported types fall back to `libdns.RR`

**API Format:**
- Requests/responses use `{"records": [...]}` wrapper object (falls back to direct array for GET responses)
- Authentication via Bearer token in Authorization header
