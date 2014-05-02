/*
https://github.com/tonnerre/golang-dns/blob/master/ex/as112/as112.go
https://github.com/ant0ine/go-json-rest
 */

package main

import (
	"github.com/codegangsta/martini"
	"runtime"
	"log"
	"net/http"
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
)

var (
	conf *goini.Ini 
	confErr error
	redisConn *redis.Client
)

type DnsRecord struct {
	Name string `json:"name"`
	Ttl uint32 `json:"ttl"`
	A string `json:"a"`
	Ns string `json:"ns"`
	Mx string `json:"mx"`
	Txt string `json:"Txt"`
	Preference uint16 `json:"preference"`
}

type DnsRecordsCollection struct {
	Pool map[string]DnsRecord
}

func (mc *DnsRecordsCollection) FromJson(jsonStr string) error {
	var data = &mc.Pool
	b := []byte(jsonStr)
	return json.Unmarshal(b, data)
}

func serve(net string) {
	err := dns.ListenAndServe(":53", net, nil)
	if err != nil {
		log.Fatal("Failed to set " + net + " listener %s\n", err.Error())
	}
}

func getRecord(name string, qType uint16) (string, error){
	typeStr, _ := dns.TypeToString[qType]
	key := conf.Str("redis", "key") + ":" + name + ":" + typeStr
	key = strings.ToLower(key)
	fmt.Println(key)
	return redisConn.Get(key).Result()
}

func setAnswer(w dns.ResponseWriter, r *dns.Msg, data []dns.RR) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Answer = data
	w.WriteMsg(m)
}

func handleZone(w dns.ResponseWriter, r *dns.Msg) {
	json, err := getRecord(r.Question[0].Name, r.Question[0].Qtype)
	records := new(DnsRecordsCollection)

	if err != nil {
		ping := redisConn.Ping()
		err := ping.Err() 

		if err != nil {
			// no connection to redis
			m := new(dns.Msg)
			m.Authoritative = true
			m.SetRcode(r, dns.RcodeServerFailure)
			w.WriteMsg(m)

			return
		}
	} else {
		err = records.FromJson(json)

		if err != nil {
			fmt.Println(err)
		}
	}

	var answer []dns.RR

	switch r.Question[0].Qtype {
	case dns.TypeA:
		for _, rec := range records.Pool {
			record := new(dns.A)
			record.Hdr = dns.RR_Header{Name: rec.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: rec.Ttl}
			record.A = net.ParseIP(rec.A)

			answer = append(answer, record)
		}

		setAnswer(w, r, answer)

	case dns.TypeNS:
		for _, rec := range records.Pool {
			record := new(dns.NS)
			record.Hdr = dns.RR_Header{Name: rec.Name, Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: rec.Ttl}
			record.Ns = rec.Ns

			answer = append(answer, record)
		}

		setAnswer(w, r, answer)

	case dns.TypeMX:
		for _, rec := range records.Pool {
			record := new(dns.MX)
			record.Hdr = dns.RR_Header{Name: rec.Name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: rec.Ttl}
			record.Preference = rec.Preference
			record.Mx = rec.Mx

			answer = append(answer, record)
		}
		
		setAnswer(w, r, answer)

	case dns.TypeTXT:
		for _, rec := range records.Pool {
			record := new(dns.TXT)
			record.Hdr = dns.RR_Header{Name: rec.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: rec.Ttl}
			record.Txt = []string{rec.Txt}

			answer = append(answer, record)
		}

		setAnswer(w, r, answer)

	default:
		m := new(dns.Msg)
		m.Authoritative = true
		m.SetRcode(r, dns.RcodeNotImplemented)
		w.WriteMsg(m)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 4)

	conf, confErr = goini.Load("config.ini")
	if confErr != nil {
        panic(confErr)
    }

    fmt.Print(conf)

    redisConn = redis.NewTCPClient(&redis.Options{
		Addr:     conf.Str("redis", "server"),
		Password: "",
		DB:       0,
	})

	defer redisConn.Close()

	dns.HandleFunc(".", handleZone)
	/*dns.HandleFunc("authors.bind.", dns.HandleAuthors)
	dns.HandleFunc("authors.server.", dns.HandleAuthors)
	dns.HandleFunc("version.bind.", dns.HandleVersion)
	dns.HandleFunc("version.server.", dns.HandleVersion)*/

	go serve("tcp")
	go serve("udp")

	go func() {
		m := martini.Classic()
		log.Fatal(http.ListenAndServe(":8080", m))	
	}()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case s := <-sig:
			log.Fatalf("Signal (%d) received, stopping\n", s)
		}
	}
}