/*
https://github.com/ant0ine/go-json-rest
http://talks.golang.org/2013/oscon-dl.slide#47
https://github.com/feiyang21687/golang/blob/160794ad61e214aff029eb84a86a18061b7144b0/groupcached/groupcached.go
 */

package main

import (
	"runtime"
	log "github.com/cihub/seelog"
	"github.com/miekg/dns"
	"os/signal"
	"syscall"
	"github.com/asjustas/goini"
	"os"
	"fmt"
	"net"
	"github.com/vmihailenco/redis/v2"
	"encoding/json"
	"strings"
	"github.com/ant0ine/go-json-rest/rest"
    "net/http"
)

var (
	conf *goini.Ini 
	confErr error
	redisConn *redis.Client
)

type Api struct {
	dnsCore *DnsCore
}

type DnsCore struct {
	cache Cache
}

type DnsRecord struct {
	Name string `json:"name"`
	Ttl uint32 `json:"ttl"`
	A string `json:"a"`
	Ns string `json:"ns"`
	Mx string `json:"mx"`
	Txt string `json:"Txt"`
	Preference uint16 `json:"preference"`
	Id int64 `json:"id"`
}

func substr(s string,pos,length int) string{
    runes:=[]rune(s)
    l := pos+length
    if l > len(runes) {
        l = len(runes)
    }
    return string(runes[pos:l])
}

func serve(net string) {
	err := dns.ListenAndServe(conf.Str("core", "listen"), net, nil)
	if err != nil {
		log.Critical(fmt.Sprintf("Failed to set " + net + " listener %s\n", err.Error()))
		os.Exit(1)
	}
}

func (core *DnsCore) loadRecords() {
	core.cache.Reset()

	_, keys, err := redisConn.Scan(0,  conf.Str("redis", "key") + ":lookup:*", 0).Result()

	if err != nil {
		log.Error(err)
		return
	}

	for _, key := range keys {
		ids, err := redisConn.LRange(key, 0, -1).Result()

	    if err != nil {
	    	log.Error(err)
	    }

	    records := []DnsRecord{}

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

		prefixLen := len(conf.Str("redis", "key") + ":lookup:")
		saveKey := substr(key, prefixLen, len(key) - prefixLen)
		core.cache.Set(saveKey, records)
	}
}

func (core *DnsCore) getRecords(name string, qType uint16) []DnsRecord {
	typeStr, _ := dns.TypeToString[qType]

	lookupKey := name + ":" + typeStr
    lookupKey = strings.ToLower(lookupKey)
    records, _ := core.cache.Get(lookupKey)

    return records
}

func (core *DnsCore) setAnswer(w dns.ResponseWriter, r *dns.Msg, data []dns.RR) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Answer = data
	w.WriteMsg(m)
}

func (core *DnsCore) handleZone(w dns.ResponseWriter, r *dns.Msg) {
	records := core.getRecords(r.Question[0].Name, r.Question[0].Qtype)

	var answer []dns.RR

	switch r.Question[0].Qtype {
	case dns.TypeA:
		for _, rec := range records {
			record := new(dns.A)
			record.Hdr = dns.RR_Header{Name: rec.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: rec.Ttl}
			record.A = net.ParseIP(rec.A)

			answer = append(answer, record)
		}

		core.setAnswer(w, r, answer)

	case dns.TypeNS:
		for _, rec := range records {
			record := new(dns.NS)
			record.Hdr = dns.RR_Header{Name: rec.Name, Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: rec.Ttl}
			record.Ns = rec.Ns

			answer = append(answer, record)
		}

		core.setAnswer(w, r, answer)

	case dns.TypeMX:
		for _, rec := range records {
			record := new(dns.MX)
			record.Hdr = dns.RR_Header{Name: rec.Name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: rec.Ttl}
			record.Preference = rec.Preference
			record.Mx = rec.Mx

			answer = append(answer, record)
		}

		core.setAnswer(w, r, answer)

	case dns.TypeTXT:
		for _, rec := range records {
			record := new(dns.TXT)
			record.Hdr = dns.RR_Header{Name: rec.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: rec.Ttl}
			record.Txt = []string{rec.Txt}

			answer = append(answer, record)
		}

		core.setAnswer(w, r, answer)

	default:
		m := new(dns.Msg)
		m.Authoritative = true
		m.SetRcode(r, dns.RcodeNotImplemented)
		w.WriteMsg(m)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 4)

	confFile := "config.ini"
	if len(os.Args) == 2 {
		confFile = os.Args[0]
	}

	conf, confErr = goini.Load(confFile)
	if confErr != nil {
        panic(confErr)
    }

	seelog := `
    <seelog>
        <outputs formatid="main">
            <console />
            <file path="` + conf.Str("core", "log") + `"/>
        </outputs>
        <formats>
            <format id="main" format="[%LEVEL] %Date %Time %Msg%n"/>
        </formats>
    </seelog>`

	logger, _ := log.LoggerFromConfigAsBytes([]byte(seelog))
	log.ReplaceLogger(logger)

	log.Info("Server started")

    redisConn = redis.NewTCPClient(&redis.Options{
		Addr:     conf.Str("redis", "server"),
		Password: "",
		DB:       0,
	})

	defer redisConn.Close()

	dnsCore := new(DnsCore)
	dnsCore.cache = Cache{store: make(map[string][]DnsRecord)}
	dnsCore.loadRecords()

	dns.HandleFunc(".", dnsCore.handleZone)
	/*dns.HandleFunc("authors.bind.", dns.HandleAuthors)
	dns.HandleFunc("authors.server.", dns.HandleAuthors)
	dns.HandleFunc("version.bind.", dns.HandleVersion)
	dns.HandleFunc("version.server.", dns.HandleVersion)*/

	go serve("tcp")
	go serve("udp")

	api := Api {
		dnsCore: dnsCore,
	}

	go func () {
		handler := rest.ResourceHandler{
	        PreRoutingMiddlewares: []rest.Middleware{
	            &rest.AuthBasicMiddleware{
	                Realm: "GoNS api",
	                Authenticator: func(userId string, password string) bool {
	                    if userId == conf.Str("api", "username") && password == conf.Str("api", "password") {
	                        return true
	                    }
	                    return false
	                },
	            },
	        },
	    }

	    handler.SetRoutes(
        	/*rest.RouteObjectMethod("GET", "/records", &api, "GetAllRecords"),*/
        	rest.RouteObjectMethod("POST", "/records", &api, "CreateRecord"),
        	/*rest.RouteObjectMethod("GET", "/records/:id", &api, "GetRecord"),
        	rest.RouteObjectMethod("PUT", "/records/:id", &api, "PutRecord"),
        	rest.RouteObjectMethod("DELETE", "/records/:id", &api, "DeleteRecord"),*/
    	)

	    http.ListenAndServe(conf.Str("api", "listen"), &handler)
	}()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case s := <-sig:
			log.Critical(fmt.Sprintf("Signal (%d) received, stopping\n", s))
			os.Exit(1)
		}
	}
}