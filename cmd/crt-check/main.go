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

// Print result of all checks.
func printResult(checkedItems *[]certexp.Check, numDays int) {
	var displayResult []string

	for _, check := range *checkedItems {
		for _, res := range check.Result {
			if res.Error.Code <= 1 && (numDays == 0 || res.Expiry.Days < numDays) {

				displayString := fmt.Sprintf("\n- Domain: %s [%s:%d]\n", check.Host.Name, res.Address, check.Host.Port)

				if res.Expiry.Days < 0 {
					displayString += fmt.Sprintf("  Expired: %d", -1*res.Expiry.Days)

					if res.Expiry.Days == -1 {
						displayString += " day ago"
					} else {
						displayString += " days ago"
					}
				} else {
					displayString += fmt.Sprintf("  Expires in: %d", res.Expiry.Days)

					if res.Expiry.Days == 1 {
						displayString += " day"
					} else {
						displayString += " days"
					}
				}

				displayString += fmt.Sprintf(":\n  * Issuer: %s\n  * Expiry: %v\n  * Subject: %v",
					res.Expiry.Issuer, res.Expiry.Date.Format(time.RFC850), res.Expiry.Subject)

				displayResult = append(displayResult, displayString)
			}
		}
	}

	if len(displayResult) > 0 {
		if numDays == 0 {
			fmt.Printf("\nAll certificates:\n")
		} else {
			fmt.Printf("\nCertificates with expiration lower than %d days:\n", numDays)
		}

		for _, item := range displayResult {
			fmt.Println(item)
		}
	} else {
		fmt.Printf("\nThere is no certificates with expiration lower than %d days.\n", numDays)
	}

	displayResult = nil

	for _, check := range *checkedItems {
		for _, res := range check.Result {
			if res.Error.Code > 1 {
				displayString := fmt.Sprintf("\n- Domain: %s", check.Host.Name)

				if res.Address != nil {
					displayString += fmt.Sprintf(" [%s:%d]", res.Address, check.Host.Port)
				}

				displayString += fmt.Sprintf("\n  Error (%d): %v", res.Error.Code, res.Error.Message)
				displayResult = append(displayResult, displayString)
			}
		}
	}

	if len(displayResult) > 0 {
		fmt.Printf("\nCheck with errors:\n")

		for _, item := range displayResult {
			fmt.Println(item)
		}
	}

	fmt.Println()
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
		fmt.Printf("Runtime error: %v\n", err)
		os.Exit(1)
	}

	for _, domain := range *domains {
		domain.Resolve(*options.CheckIPv6)

		for _, addr := range domain.Addresses {
			var check *certexp.Check = certexp.NewCheck(certexp.HostInfo{Name: domain.Name, Address: addr, Port: domain.Port})

			if err := check.Process(); err != nil {
				fmt.Printf("Expiration check error: %v\n", err)
			}

			checkedItems = append(checkedItems, *check)
		}
	}

	printResult(&checkedItems, *numDays)

	os.Exit(0)
}
