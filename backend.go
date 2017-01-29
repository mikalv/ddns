package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// This type implements the PowerDNS-Pipe-Backend protocol and generates
// the response data if possible
type PowerDnsBackend struct {
	hosts HostBackend
}

func NewPowerDnsBackend(backend HostBackend) *PowerDnsBackend {
	return &PowerDnsBackend{
		hosts: backend,
	}
}

func (b *PowerDnsBackend) Run() {
	bio := bufio.NewReader(os.Stdin)

	// handshake with PowerDNS
	_, _, _ = bio.ReadLine()
	fmt.Println("OK\tDDNS Go Backend")

	for {
		line, _, err := bio.ReadLine()
		if err != nil {
			fmt.Println("FAIL")
			continue
		}

		if err = b.HandleRequest(string(line)); err != nil {
			fmt.Printf("LOG\t'%s'\n", err)
		}

		fmt.Println("END")
	}
}

func (b *PowerDnsBackend) HandleRequest(line string) error {
	if Verbose {
		fmt.Printf("LOG\t'%s'\n", line)
	}

	parts := strings.Split(line, "\t")
	if len(parts) != 6 {
		return errors.New("Invalid line")
	}

	query_name := parts[1]
	query_class := parts[2]
	query_type := parts[3]
	query_id := parts[4]

	var response, record string
	record = query_type

	switch query_type {
	case "SOA":
		response = fmt.Sprintf("%s. hostmaster.example.com. %d 1800 3600 7200 5",
			DdnsSoaFqdn, b.getSoaSerial())

	case "NS":
		response = fmt.Sprintf("%s.", DdnsSoaFqdn)

	case "A", "ANY":
		// get the host part of the fqdn: pi.d.example.org -> pi
		hostname := ""
		if strings.HasSuffix(query_name, DdnsDomain) {
			hostname = query_name[:len(query_name)-len(DdnsDomain)]
		}

		if hostname == "" {
			return nil
		}

		var err error
		var host *Host

		if host, err = b.hosts.GetHost(hostname); err != nil {
			return err
		}

		response = host.Ip

		record = "A"
		if !host.IsIPv4() {
			record = "AAAA"
		}

	default:
		return nil
	}

	fmt.Printf("DATA\t%s\t%s\t%s\t10\t%s\t%s\n",
		query_name, query_class, record, query_id, response)

	return nil
}

func getSoaSerial() int64 {
	// return current time in seconds
	return time.Now().Unix()
}
