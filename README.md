## GoNS
[![Gobuild Download](http://gobuild.io/badge/github.com/asjustas/gons/download.png)](http://gobuild.io/github.com/asjustas/gons)

GoNS is experimental name server in golang.

## Install from binary
You're free to download up-to-date binary from [gobuild.io](http://gobuild.io/download/github.com/asjustas/gons).

## HTTP API
Best way to configure GoNS server

### Authentication
You must use HTTP Basic authentication. Username and password is set in configuration file.

### Handling errors
If GoNS is having trouble, you might see a 5xx error. 500 means that the app is entirely down, but you might also see ```502 Bad Gateway```, ```503 Service Unavailable```, or ```504 Gateway Timeout```. It's your responsibility in all of these cases to retry your request later.

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

```json
{
    "type": "cname",
    "name": "git.u3.lt",
    "ttl": 3600,
    "cname": "bitbucket.org."
    ...
}
```

This will return ```201 Created```, with the current JSON representation of the record if the creation was a success.

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
    }
]
```