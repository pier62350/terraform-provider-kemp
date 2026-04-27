---
name: kemp-api-lookup
description: Query the official Kemp LoadMaster API v2 Postman collection for a specific /accessv2 command. Use whenever a user asks for "what fields does <cmd> take", "the body for <cmd>", "look up <cmd> in the spec", or before implementing/extending a resource that hits a new Kemp command.
---

# Looking up a Kemp /accessv2 command

The official spec is a published Postman collection cached locally at
`~/kemp-api-spec/collection.json` (≈2.2 MB, 453 commands across 27
categories). **This is the source of truth — prefer it over any
third-party client library or reverse-engineering from a running unit.**

## Fast path: query for one command

```bash
python3 - <<'EOF'
import json, re, sys
TARGET = "addvs"   # ← replace with the command name

with open('/home/pierre/kemp-api-spec/collection.json') as f:
    c = json.load(f)

def walk(items, path=''):
    for it in items:
        if 'item' in it:
            yield from walk(it['item'], path + '/' + it.get('name',''))
        else:
            yield path + '/' + it.get('name',''), it

found = 0
for path, req in walk(c['item']):
    body = req.get('request',{}).get('body',{}).get('raw','')
    m = re.search(r'"cmd"\s*:\s*"' + re.escape(TARGET) + r'"', body)
    if not m:
        continue
    found += 1
    print(f"\n=== {path} ===")
    print("REQUEST BODY:")
    print(body)
    desc = req.get('request',{}).get('description','')
    clean = re.sub(r'<[^>]+>', ' ', desc)
    clean = re.sub(r'\s+', ' ', clean).strip()
    if clean:
        print("\nDESCRIPTION:", clean[:1500])
    for r in req.get('response', [])[:1]:
        rb = r.get('body') or ''
        if rb:
            print("\nRESPONSE EXAMPLE:", rb[:600])

if found == 0:
    print(f"No matches for cmd={TARGET!r}")
EOF
```

## Search for multiple related commands

When designing a new resource family, list every related command first:

```bash
python3 - <<'EOF'
import json, re
PATTERN = r"acme"   # regex matched against cmd values

with open('/home/pierre/kemp-api-spec/collection.json') as f:
    c = json.load(f)
def walk(items, path=''):
    for it in items:
        if 'item' in it:
            yield from walk(it['item'], path + '/' + it.get('name',''))
        else:
            yield path + '/' + it.get('name',''), it

for path, req in walk(c['item']):
    body = req.get('request',{}).get('body',{}).get('raw','')
    m = re.search(r'"cmd"\s*:\s*"(\w*' + PATTERN + r'\w*)"', body, re.IGNORECASE)
    if m:
        print(f"  CMD={m.group(1):25s}  PATH={path}")
EOF
```

## When the cached spec is stale

Re-fetch from the Postman-hosted source:

```bash
curl -sL "https://loadmasterapiv2.docs.progress.com/api/collections/1897577/VUjPK5sb?segregateAuth=true&versionTag=latest" \
  -o ~/kemp-api-spec/collection.json
```

The `1897577/VUjPK5sb` IDs are stable across spec updates (they're embedded
in the published-docs site's HTML at the `<meta name="collectionId">` and
`<meta name="publishedId">` tags).

## Reading the result

For each matched command you get:

- **REQUEST BODY** — the exact JSON shape Kemp expects, including every
  optional parameter Kemp's own examples use. Field names matter — Kemp
  is fussy (see CLAUDE.md "Wire-format quirks").
- **DESCRIPTION** — HTML-stripped doc text describing each parameter, its
  type, default, and valid range. For commands with rich enums (rule
  types, intercept modes), this is where the enum values are spelled out.
- **RESPONSE EXAMPLE** — the response shape, including which keys the
  response wraps in (`Data: {...}`, top-level keys, lists, etc.). Use
  this to design the Go response struct's `json:"…"` tags.

## After looking it up

When implementing the matching Go client method:

1. Add it to `internal/loadmaster/<area>.go` following the
   `c.call(ctx, "<cmd>", body, &resp)` pattern.
2. Use `json:"PascalName"` tags on every response struct field — Kemp's
   wire format is PascalCase, not snake_case or anything Go-idiomatic.
3. Use bare numeric Index for VS references; `!N` for RS references; the
   `addrs` command takes `rsport` (not `port`) for the RS port.
4. Cross-check with an actual call against a running LoadMaster before
   shipping — spec examples occasionally lag firmware.
