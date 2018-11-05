package cfddns

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

// Daemon is used to monitor.
func Daemon(ctx context.Context, zoneID string, api *cloudflare.API, dur time.Duration) {
	ticker := time.NewTicker(dur)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			recs, err := api.DNSRecords(zoneID, cloudflare.DNSRecord{})
			if err != nil {
				log.Println(err)
				continue
			}
			for _, rec := range recs {
				switch rec.Type {
				case "A":
					curIP, err := GetIPV4()
					if err != nil {
						continue
					}
					if curIP == rec.Content {
						continue
					}
					rec.Content = curIP
					if err := api.UpdateDNSRecord(zoneID, rec.ID, rec); err != nil {
						log.Println(err)
					}
					log.Printf("Updated IP address to %s\n", curIP)
				default:
					continue
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

// GetIPV4 returns an public IPV4.
func GetIPV4() (string, error) {
	// TODO: add more robust checks
	return checkDynDNSIPV4()
}

// checkDynDNSIPV4 checks dyndns for the current hosts IP address.
func checkDynDNSIPV4() (string, error) {
	res, err := http.Get("http://checkip.dyndns.org/")
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	ipBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	ipStrs := strings.Split(string(ipBytes), "<body>")
	if len(ipStrs) < 1 {
		return "", fmt.Errorf("No IP returned")
	}
	ipStrs = strings.Split(ipStrs[1], "</body>")
	ipStrs = strings.Split(ipStrs[0], ":")

	return strings.TrimPrefix(ipStrs[1], " "), nil
}

// check
// http://ipdetect.dnspark.com/

// check
// http://dns.loopia.se/checkip/checkip.php
