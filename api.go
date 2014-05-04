package main

import (
    "github.com/ant0ine/go-json-rest/rest"
    "net/http"
    "strings"
    "encoding/json"
    "strconv"
)

type ApiDnsRecord struct {
    Type string
    Name string
    Ttl uint32
    A string
    Ns string
    Mx string
    Txt string
    Preference uint16
}

func (a *Api) CreateRecord(w rest.ResponseWriter, r *rest.Request) {
    record := ApiDnsRecord{}
    err := r.DecodeJsonPayload(&record)

    if err != nil {
        rest.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    redisRec := &DnsRecord{
        Name: record.Name,
        Ttl: record.Ttl,
        A: record.A,
        Ns: record.Ns,
        Mx: record.Mx,
        Txt: record.Txt,
        Preference: record.Preference,
    }

    idInt, err := redisConn.Incr(conf.Str("redis", "key") + ":counters:ids").Result()
    if err != nil {
        rest.Error(w, "a"+err.Error(), http.StatusInternalServerError)
        return
    }

    id := strconv.FormatInt(idInt, 10)

    key := conf.Str("redis", "key") + ":records:" + id
    key = strings.ToLower(key)

    redisRecJson, err := json.Marshal(redisRec)
    if err != nil {
        rest.Error(w, err.Error(), http.StatusInternalServerError)
        return
    } 

    err = redisConn.Set(key, string(redisRecJson)).Err()
    if err != nil {
        rest.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    lookupKey := conf.Str("redis", "key") + ":lookup:" + record.Name + ":" + record.Type
    lookupKey = strings.ToLower(lookupKey)
    _, err = redisConn.RPush(lookupKey, id).Result()
    if err != nil {
        rest.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    redisRec.Id = idInt
    w.WriteJson(&redisRec)
}