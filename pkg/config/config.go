package config

import (
	"crt-mon/pkg/certexp"
	"flag"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Options struct {
	CheckIPv6  *bool
	ConfigFile *string
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) CommonFlags() {
	o.CheckIPv6 = flag.Bool("6", false, "Check IPv6")
	o.ConfigFile = flag.String("file", "/opt/etc/crt-hosts.yml", "File contains hosts list")
}

func Parse(configFile string) (*[]certexp.Domain, error) {
	var domains []certexp.Domain

	data, err := os.ReadFile(configFile)

	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &domains); err != nil {
		log.Printf("Config error: %v", err)
		return nil, err
	}

	for i := 0; i < len(domains); i++ {
		if domains[i].Port == 0 {
			domains[i].Port = 443
		}
	}

	return &domains, nil
}
