package main

import (
	"encoding/json"
	"github.com/ant0ine/go-json-rest/rest"
	log "github.com/cihub/seelog"
	"net/http"
	"strconv"
	"strings"
)

type ApiDnsRecord struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	Ttl        uint32 `json:"ttl"`
	A          string `json:"a"`
	AAAA       string `json:"aaaa"`
	Ns         string `json:"ns"`
	Mx         string `json:"mx"`
	Txt        string `json:"txt"`
	Cname      string `json:"cname"`
	Preference uint16 `json:"preference"`
	Refresh    uint32 `json:"refresh"`
	Retry      uint32 `json:"retry"`
	Expire     uint32 `json:"expire"`
	Minttl     uint32 `json:"minttl"`
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
		rest.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	id := strconv.FormatInt(idInt, 10)

	redisRec := &DnsRecord{
		Id:         idInt,
		Type:       record.Type,
		Name:       record.Name,
		A:          record.A,
		AAAA:       record.AAAA,
		Ns:         record.Ns,
		Mx:         record.Mx,
		Txt:        record.Txt,
		Cname:      record.Cname,
		Preference: record.Preference,
		Ttl:        record.Ttl,
		Refresh:    record.Refresh,
		Retry:      record.Retry,
		Expire:     record.Expire,
		Minttl:     record.Minttl,
	}

	key := conf.Str("redis", "key") + ":records:" + id
	key = strings.ToLower(key)

	redisRecJson, err := json.Marshal(redisRec)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	err = redisConn.Set(key, string(redisRecJson)).Err()
	if err != nil {
		rest.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	lookupKey := conf.Str("redis", "key") + ":lookup:" + record.Name + ":" + record.Type
	lookupKey = strings.ToLower(lookupKey)
	_, err = redisConn.RPush(lookupKey, id).Result()
	if err != nil {
		rest.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	redisRec.Id = idInt

	w.WriteHeader(http.StatusCreated)
	w.WriteJson(&redisRec)

	api.dnsCore.loadRecords()
}

func (api *Api) GetRecord(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")

	key := conf.Str("redis", "key") + ":records:" + id
	jsonStr, err := redisConn.Get(key).Result()

	if err != nil {
		log.Error(err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		record := DnsRecord{}
		if err := json.Unmarshal([]byte(jsonStr), &record); err != nil {
			log.Error(err)
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteJson(&record)
	}
}

func (api *Api) GetAllRecords(w rest.ResponseWriter, r *rest.Request) {
	keys, err := redisConn.Keys(conf.Str("redis", "key") + ":lookup:*").Result()

	if err != nil {
		log.Error(err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	records := []DnsRecord{}

	for _, key := range keys {
		ids, err := redisConn.LRange(key, 0, -1).Result()

		if err != nil {
			log.Error(err)
			rest.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, id := range ids {
			key := conf.Str("redis", "key") + ":records:" + id
			jsonStr, err := redisConn.Get(key).Result()

			if err != nil {
				log.Error(err)
				rest.Error(w, err.Error(), http.StatusInternalServerError)
				return
			} else {
				record := DnsRecord{}
				if err := json.Unmarshal([]byte(jsonStr), &record); err != nil {
					log.Error(err)
					rest.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				records = append(records, record)
			}
		}
	}

	w.WriteJson(&records)
}
