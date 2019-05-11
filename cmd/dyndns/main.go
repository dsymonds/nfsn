/*
dyndns updates DNS records on NFSN to match your public IP.
*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dsymonds/nfsn"
)

var (
	domain    = flag.String("domain", "", "which domain to modify")
	subDomain = flag.String("subdomain", "", "which subdomain tracks the public IP")
	q         = flag.Bool("q", false, "whether to be quiet, except for errors")
)

func infof(format string, args ...interface{}) {
	if *q {
		return
	}
	log.Printf(format, args...)
}

func main() {
	// TODO: better usage
	flag.Parse()
	if *domain == "" {
		log.Fatal("-domain must be set")
	}
	if *subDomain == "" {
		log.Fatal("-subdomain must be set")
	}

	client, err := nfsn.NewClient()
	if err != nil {
		log.Fatalf("Initialising: %v", err)
	}

	ip, err := publicIP()
	if err != nil {
		log.Fatalf("Getting public IP: %v", err)
	}
	infof("Public IP: %s", ip)

	dns := client.DNS(*domain)
	rrs, err := dns.RRs()
	if err != nil {
		log.Fatalf("Getting current RRs for %s: %v", *domain, err)
	}
	existing, ok := nfsn.DNSRR{}, false
	for _, rr := range rrs {
		if rr.Name == *subDomain {
			infof("Found existing RR: %v", rr)
			existing, ok = rr, true
			break
		}
	}
	if ok {
		if existing.Data == ip && existing.Type == "A" {
			infof("(%s).%s already configured correctly", *subDomain, *domain)
			return
		}
		log.Printf("Existing RR has incorrect configuration")
		if err := dns.DeleteRR(existing); err != nil {
			log.Fatalf("Deleting existing RR: %v", err)
		}
	}

	if err := dns.AddRR(nfsn.DNSRR{
		Name: *subDomain,
		Type: "A",
		Data: ip,
		TTL:  300, // 5m
	}); err != nil {
		log.Fatalf("Setting new RR: %v", err)
	}
	infof("New RR created")
}

func publicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("reading response: %v", err)
	}
	var out struct {
		IP string `json:"ip"`
	}
	err = json.Unmarshal(body, &out)
	return out.IP, err
}
