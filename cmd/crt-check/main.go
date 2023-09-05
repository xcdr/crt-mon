package main

import (
	"crt-mon/pkg/certexp"
	"crt-mon/pkg/config"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	version = "dev"
	build   = "none"
	author  = "undefined"
)

func printResult(checkedItems *[]certexp.Check, numDays int) {
	if numDays == 0 {
		fmt.Printf("\nAll certificates:\n")
	} else {
		fmt.Printf("\nCertificates with expiration lower than %d days:\n", numDays)
	}

	for _, check := range *checkedItems {
		for _, res := range check.Result {
			if res.Error.Code <= 1 && (numDays == 0 || res.Expiry.Days < numDays) {
				fmt.Printf("\n- Domain: %s [%s:%d]\n", check.Host.Name, res.Address, check.Host.Port)
				if res.Expiry.Days < 0 {
					fmt.Printf("  Expired: %d", -1*res.Expiry.Days)

					if res.Expiry.Days == -1 {
						fmt.Print(" day ago")
					} else {
						fmt.Print(" days ago")
					}
				} else {
					fmt.Printf("  Expires in: %d", res.Expiry.Days)

					if res.Expiry.Days == 1 {
						fmt.Print(" day")
					} else {
						fmt.Print(" days")
					}
				}
				fmt.Printf(":\n  * Issuer: %s\n  * Expiry: %v\n  * Subject: %v\n",
					res.Expiry.Issuer, res.Expiry.Date.Format(time.RFC850), res.Expiry.Subject)
			}

			if res.Error.Code > 1 {
				fmt.Printf("\n- Domain: %s [%s:%d]\n", check.Host.Name, res.Address, check.Host.Port)
				fmt.Printf("  Error: %v\n", res.Error.Message)
			}
		}
	}
}

func main() {
	var checkedItems []certexp.Check

	options := config.NewOptions()
	options.CommonFlags()

	numDays := flag.Int("days", 0, "Number of days for notification")

	flag.Parse()

	program := filepath.Base(os.Args[0])

	fmt.Printf("%s started, version: %s+%s, author: %s\n", program, version, build, author)

	domains, err := config.Parse(*options.ConfigFile)

	if err != nil {
		fmt.Printf("Unexpected error: %v\n", err)
		os.Exit(1)
	}

	for _, domain := range *domains {
		domain.Resolve(*options.CheckIPv6)

		for _, addr := range domain.Addresses {
			var check *certexp.Check = certexp.NewCheck(certexp.HostInfo{Name: domain.Name, Address: addr, Port: domain.Port})

			if err := check.Expiration(); err != nil {
				fmt.Printf("Expiration check error: %v\n", err)
			}

			checkedItems = append(checkedItems, *check)
		}
	}

	printResult(&checkedItems, *numDays)

	os.Exit(0)
}
