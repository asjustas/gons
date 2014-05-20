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
| name | Record name
| ttl | Record ttl value
| a |
| aaaa |
| ns |
| mx |
| txt |
| cname |
| preference |

### Get record
```GET /v1/records/<id>.json```

Retrieve the particular record by id

### Get all records
```GET /v1/records.json```

Retrieve the existing records