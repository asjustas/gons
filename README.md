## GoNS
[![Gobuild Download](http://gobuild.io/badge/github.com/asjustas/gons/download.png)](http://gobuild.io/github.com/asjustas/gons)

GoNS is experimental name server in golang.

## Install from binary
You're free to download up-to-date binary from [gobuild.io](http://gobuild.io/download/github.com/asjustas/gons).

## HTTP API
Best way to configure GoNS server

### Create record
```POST /v1/records.json```

Add dns record

| Parameter     | Description   |
| ------------- |---------------|
| type | Record type |
| name | Record name |
| ttl | Record ttl value |
| a | Record a value |
| aaaa | Record aaaa value |
| ns | Record ns value |
| mx | Record mx value |
| txt | Record txt value |
| cname | Record cname value |
| preference | Record prefecence value |

### Get record
```GET /v1/records/<id>.json```

Retrieve the particular record by id

```json
{
    "id": 6,
    "type": "a",
    "name": "ui8.lt.",
    "a": "8.8.8.8",
    "aaaa": "",
    "ns": "",
    "mx": "",
    "txt": "",
    "cname": "",
    "preference": 0,
    "ttl": 3600
}
```

### Get all records
```GET /v1/records.json```

Retrieve the existing records

```json
[
    {
        "id": 7,
        "type": "a",
        "name": "ui8.lt.",
        "a": "8.8.8.8",
        "aaaa": "",
        "ns": "",
        "mx": "",
        "txt": "",
        "cname": "",
        "preference": 0,
        "ttl": 3600
    },
    {
        "id": 8,
        "type": "a",
        "name": "ns2.ui8.lt.",
        "a": "127.0.0.2",
        "aaaa": "",
        "ns": "",
        "mx": "",
        "txt": "",
        "cname": "",
        "preference": 0,
        "ttl": 3600
    },
    {
        "id": 9,
        "type": "a",
        "name": "ns1.ui8.lt.",
        "a": "127.0.0.1",
        "aaaa": "",
        "ns": "",
        "mx": "",
        "txt": "",
        "cname": "",
        "preference": 0,
        "ttl": 3600
    }
]
```