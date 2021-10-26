package config

import (
	"bufio"
	"crt-mon/pkg/certexp"
	"flag"
	"os"
	"strings"
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
	o.ConfigFile = flag.String("file", "/opt/etc/crt-hosts.conf", "File contains hosts list")
}

func Parse(configFile string) (*[]certexp.HostInfo, error) {
	var hosts []certexp.HostInfo

	file, err := os.Open(configFile)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if err := scanner.Err(); err != nil {
			return nil, err
		}

		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		hosts = append(hosts, certexp.HostInfo{Name: line, Port: 443})
	}

	return &hosts, nil
}
