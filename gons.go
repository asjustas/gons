/*
https://github.com/tonnerre/golang-dns/blob/master/ex/as112/as112.go
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
)

func MakeRR(s string) dns.RR { 
	r, _ := dns.NewRR(s); return r
}

func handleZone(w dns.ResponseWriter, r *dns.Msg) {
	switch r.Question[0].Qtype {
	case dns.TypeA:
		record := "jv.lt. IN A 127.0.0.1\n"
		rr := MakeRR(record)
		rrx := rr.(*dns.A)

		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = true
		m.Answer = []dns.RR{rrx}
		w.WriteMsg(m)

	case dns.TypeNS:
		record := new(dns.NS)
		record.Hdr = dns.RR_Header{Name: "jv.lt.", Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: 3600}
		record.Ns = "ns.jv.lt."

		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = true
		m.Answer = []dns.RR{record}
		w.WriteMsg(m)

	case dns.TypeMX:
		record := new(dns.MX)
		record.Hdr = dns.RR_Header{Name: "jv.lt.", Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: 3600}
		record.Preference = 10
		record.Mx = "mx.jv.lt."

		record2 := new(dns.MX)
		record2.Hdr = dns.RR_Header{Name: "jv.lt.", Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: 3600}
		record2.Preference = 20
		record2.Mx = "mx2.jv.lt."

		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = true
		m.Answer = []dns.RR{record, record2}
		w.WriteMsg(m)

	case dns.TypeTXT:
		record := new(dns.TXT)
		record.Hdr = dns.RR_Header{Name: "jv.lt.", Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 3600}
		record.Txt = []string{"google-site-verification=rXOxyZounnZasA8Z7oaD3c14JdjS9aKSWvsR1EbUSIQ"}

		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = true
		m.Answer = []dns.RR{record}
		w.WriteMsg(m)

	default:
		m := new(dns.Msg)
		m.Authoritative = true
		m.SetRcode(r, dns.RcodeNotImplemented)
		w.WriteMsg(m)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 4)

	conf, err := goini.Load("config.ini")

	if err != nil {
        panic(err)
    }

    fmt.Print(conf)

	dns.HandleFunc("jv.lt.", handleZone)
	/*dns.HandleFunc("authors.bind.", dns.HandleAuthors)
	dns.HandleFunc("authors.server.", dns.HandleAuthors)
	dns.HandleFunc("version.bind.", dns.HandleVersion)
	dns.HandleFunc("version.server.", dns.HandleVersion)*/

	go func() {
		err := dns.ListenAndServe(":53", "tcp", nil)
		if err != nil {
			log.Fatal("Failed to set tcp listener %s\n", err.Error())
		}
	}()

	go func() {
		err := dns.ListenAndServe(":53", "udp", nil)
		if err != nil {
			log.Fatal("Failed to set udp listener %s\n", err.Error())
		}
	}()

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