package main

import (
    "github.com/ant0ine/go-json-rest/rest"
    "net/http"
    "strings"
    "encoding/json"
    "strconv"
    log "github.com/cihub/seelog"
)

type ApiDnsRecord struct {
    Type string `json:"type"`
    Name string `json:"name"`
    Ttl uint32 `json:"ttl"`
    A string `json:"a"`
    AAAA string `json:"aaaa"`
    Ns string `json:"ns"`
    Mx string `json:"mx"`
    Txt string `json:"txt"`
    Cname string `json:"cname"`
    Preference uint16 `json:"preference"`
}

func (api *Api) CreateRecord(w rest.ResponseWriter, r *rest.Request) {
    record := ApiDnsRecord{}
    err := r.DecodeJsonPayload(&record)

    if err != nil {
        rest.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    idInt, err := redisConn.Incr(conf.Str("redis", "key") + ":counters:ids").Result()
    if err != nil {
        rest.Error(w, "a"+err.Error(), http.StatusInternalServerError)
        return
    }

    id := strconv.FormatInt(idInt, 10)

    redisRec := &DnsRecord{
        Id: idInt,
        Type: record.Type,
        Name: record.Name,
        A: record.A,
        Ns: record.Ns,
        Mx: record.Mx,
        Txt: record.Txt,
        Preference: record.Preference,
        Ttl: record.Ttl,
    }

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

    api.dnsCore.loadRecords()
}

func (api *Api) GetAllRecords(w rest.ResponseWriter, r *rest.Request) {
    _, keys, err := redisConn.Scan(0,  conf.Str("redis", "key") + ":lookup:*", 0).Result()

    if err != nil {
        log.Error(err)
        return
    }

    records := []DnsRecord{}

    for _, key := range keys {
        ids, err := redisConn.LRange(key, 0, -1).Result()

        if err != nil {
            log.Error(err)
        }

        for _, id := range ids {
            key := conf.Str("redis", "key") + ":records:" + id
            jsonStr, err := redisConn.Get(key).Result()

            if err != nil {
                log.Error(err)
            } else {
                record := DnsRecord{}
                if err := json.Unmarshal([]byte(jsonStr), &record); err != nil {
                    panic(err)
                }

                records = append(records, record)   
            }
        }
    }

    w.WriteJson(&records)
}