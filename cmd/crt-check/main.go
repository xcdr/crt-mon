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
		fmt.Printf("\nChecked hosts:\n")
	} else {
		fmt.Printf("\nHosts with expiration lower than %d days:\n", numDays)
	}

	for _, check := range *checkedItems {
		for _, res := range check.Result {
			if res.Error.Code <= 1 && (numDays == 0 || res.Expiry.Days < numDays) {
				fmt.Printf("\n- %s:%d [%s]\n", check.Host.Name, check.Host.Port, res.Address)
				fmt.Printf("\n  Issuer: %s\n  Expiry: %v\n  Subject: %v\n  Expires: %d\n",
					res.Expiry.Issuer, res.Expiry.Date.Format(time.RFC850), res.Expiry.Subject, res.Expiry.Days)
			}

			if res.Error.Code > 1 {
				fmt.Printf("\n- %s:%d [%s]\n", check.Host.Name, check.Host.Port, res.Address)
				fmt.Printf("\n  Error: %v\n", res.Error.Message)
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

	hosts, err := config.Parse(*options.ConfigFile)

	if err != nil {
		fmt.Printf("Unexpected error: %v\n", err)
		os.Exit(1)
	}

	for _, host := range *hosts {
		var check *certexp.Check = certexp.NewCheck(host)

		if err := check.Expiration(*options.CheckIPv6); err != nil {
			fmt.Printf("Expiration check error: %v\n", err)
		}

		checkedItems = append(checkedItems, *check)
	}

	printResult(&checkedItems, *numDays)

	os.Exit(0)

}
