/*
dyndns updates DNS records on NFSN to match your public IP.
*/
package main

import (
	"flag"
	"log"

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
	rrs, err := client.DNSListRRs(*domain)
	if err != nil {
		log.Fatalf("Getting current RRs for %s: %v", *domain, err)
	}
	log.Printf("Found %d existing RRs:", len(rrs))
	for _, rr := range rrs {
		log.Printf("* %#v", rr)
	}
}
