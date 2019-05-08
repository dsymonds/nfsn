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
)

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
	log.Printf("Public IP: %s", ip)

	dns := client.DNS(*domain)
	rrs, err := dns.RRs()
	if err != nil {
		log.Fatalf("Getting current RRs for %s: %v", *domain, err)
	}
	existing := ""
	for _, rr := range rrs {
		if rr.Name == *subDomain && rr.Type == "A" {
			log.Printf("Found existing RR: %#v", rr)
			existing = rr.Data
			break
		}
	}
	if existing == ip {
		log.Printf("(%s).%s already set correctly", *subDomain, *domain)
		return
	}
	if existing != "" {
		// TODO: delete existing record.
	}

	if err := dns.AddRR(nfsn.DNSRR{
		Name: *subDomain,
		Type: "A",
		Data: ip,
		TTL:  300, // 5m
	}); err != nil {
		log.Fatalf("Setting new RR: %v", err)
	}
	log.Printf("New RR created")
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
