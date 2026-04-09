// Command netdata prints network information from the local system.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"forge/tmp/nettest/pkg/dns"
	"forge/tmp/nettest/pkg/interfaces"
	"forge/tmp/nettest/pkg/ip"
	"forge/tmp/nettest/pkg/routes"
)

func main() {
	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)

	var help bool
	var showIfaces, showRoutes, showPublic, showCurrent, showDNS bool
	flag.BoolVar(&help, "h", false, "show help and exit")
	flag.BoolVar(&help, "help", false, "show help and exit")
	flag.BoolVar(&showIfaces, "i", false, "list all network interfaces")
	flag.BoolVar(&showIfaces, "interfaces", false, "list all network interfaces")
	flag.BoolVar(&showRoutes, "r", false, "display routing table")
	flag.BoolVar(&showRoutes, "routes", false, "display routing table")
	flag.BoolVar(&showPublic, "p", false, "show public IP address")
	flag.BoolVar(&showPublic, "public-ip", false, "show public IP address")
	flag.BoolVar(&showCurrent, "c", false, "show current IP address of each interface")
	flag.BoolVar(&showCurrent, "current-ip", false, "show current IP address of each interface")
	flag.BoolVar(&showDNS, "d", false, "display DNS resolution information")
	flag.BoolVar(&showDNS, "dns", false, "display DNS resolution information")
	usage := func() {
		w := flag.CommandLine.Output()
		fmt.Fprintln(w, "netdata prints local network information (interfaces, routes, IPs, DNS).")
		fmt.Fprintln(w, "With no options, all sections are printed. Use -i, -r, -p, -c, and -d to limit output.")
		fmt.Fprintf(w, "\nUsage: %s [options]\n\nOptions:\n", os.Args[0])
		flag.CommandLine.PrintDefaults()
	}
	flag.Usage = usage
	flag.CommandLine.Usage = usage

	out := flag.CommandLine.Output()
	flag.CommandLine.SetOutput(io.Discard)
	parseErr := flag.CommandLine.Parse(os.Args[1:])
	flag.CommandLine.SetOutput(out)
	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n\n", parseErr)
		fmt.Fprintln(os.Stderr, "Try -h or --help for valid options.")
		flag.Usage()
		os.Exit(2)
	}

	if help {
		flag.Usage()
		os.Exit(0)
	}

	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "error: unexpected arguments %q — netdata only accepts flags, not positional arguments.\n\n", flag.Args())
		fmt.Fprintln(os.Stderr, "Try -h or --help for valid options.")
		flag.Usage()
		os.Exit(2)
	}

	if !showIfaces && !showRoutes && !showPublic && !showCurrent && !showDNS {
		showIfaces, showRoutes, showPublic, showCurrent, showDNS = true, true, true, true, true
	}

	sep := false
	if showIfaces {
		if err := printInterfaces(); err != nil {
			fmt.Fprintf(os.Stderr, "interfaces: %v\n", err)
			os.Exit(1)
		}
		sep = true
	}
	if showRoutes {
		if sep {
			fmt.Println()
		}
		if err := printRoutes(); err != nil {
			fmt.Fprintf(os.Stderr, "routes: %v\n", err)
			os.Exit(1)
		}
		sep = true
	}
	if showPublic {
		if sep {
			fmt.Println()
		}
		if err := printPublicIP(); err != nil {
			fmt.Fprintf(os.Stderr, "public ip: %v\n", err)
			os.Exit(1)
		}
		sep = true
	}
	if showCurrent {
		if sep {
			fmt.Println()
		}
		if err := printCurrentIPs(); err != nil {
			fmt.Fprintf(os.Stderr, "current ip: %v\n", err)
			os.Exit(1)
		}
		sep = true
	}
	if showDNS {
		if sep {
			fmt.Println()
		}
		if err := printDNS(); err != nil {
			fmt.Fprintf(os.Stderr, "dns: %v\n", err)
			os.Exit(1)
		}
	}
}

func printInterfaces() error {
	list, err := interfaces.List()
	if err != nil {
		return err
	}
	for _, ifi := range list {
		fmt.Printf("Interface: %s (index %d)\n", ifi.Name, ifi.Index)
		fmt.Printf("  MTU: %d  Flags: %s\n", ifi.MTU, ifi.Flags.String())
		if ifi.HWAddr != "" {
			fmt.Printf("  Hardware: %s\n", ifi.HWAddr)
		}
		if len(ifi.Addrs) == 0 {
			fmt.Println("  Addresses: (none)")
			continue
		}
		fmt.Println("  Addresses:")
		for _, a := range ifi.Addrs {
			fmt.Printf("    %s\n", a)
		}
	}
	return nil
}

func printRoutes() error {
	list, err := routes.List()
	if err != nil {
		return err
	}
	fmt.Println("Kernel IP routing table")
	// Column layout matches net-tools `route` / `route -n`.
	const row = "%-16s %-16s %-16s %-5s %6s %5s %7s %s\n"
	fmt.Printf(row, "Destination", "Gateway", "Genmask", "Flags", "Metric", "Ref", "Use", "Iface")
	for _, r := range list {
		dest := routes.FormatDestination(r.Destination, r.Mask)
		flg := routes.FormatRTFFlags(r.Flags)
		fmt.Printf(row, dest, r.Gateway, r.Mask, flg, r.Metric, r.RefCnt, r.Use, r.Iface)
	}
	return nil
}

func printPublicIP() error {
	addr, err := ip.Public()
	if err != nil {
		return err
	}
	fmt.Println("Public IP")
	fmt.Printf("  %s\n", addr)
	return nil
}

func printCurrentIPs() error {
	list, err := ip.ListCurrent()
	if err != nil {
		return err
	}
	if len(list) == 0 {
		fmt.Println("(no addresses)")
		return nil
	}
	for _, row := range list {
		fmt.Printf("Interface: %s\n", row.Name)
		for _, a := range row.Addrs {
			fmt.Printf("  %s\n", a)
		}
	}
	return nil
}

func printDNS() error {
	info, err := dns.LoadSystem()
	if err != nil {
		return err
	}
	fmt.Println("DNS configuration (/etc/resolv.conf)")
	if len(info.Nameservers) == 0 {
		fmt.Println("  Nameservers: (none)")
	} else {
		fmt.Println("  Nameservers (libc / applications):")
		for _, ns := range info.Nameservers {
			fmt.Printf("    %s\n", ns)
		}
	}
	if len(info.RecursiveResolvers) > 0 {
		fmt.Println("Recursive resolvers in use (systemd-resolved, /run/systemd/resolve/resolv.conf):")
		for _, ns := range info.RecursiveResolvers {
			fmt.Printf("    %s\n", ns)
		}
	}
	if info.Domain != "" {
		fmt.Printf("  Domain: %s\n", info.Domain)
	}
	if len(info.Search) > 0 {
		fmt.Println("  Search list:")
		for _, s := range info.Search {
			fmt.Printf("    %s\n", s)
		}
	}
	if len(info.Options) > 0 {
		fmt.Println("  Options:")
		for _, o := range info.Options {
			fmt.Printf("    %s\n", o)
		}
	}
	return nil
}
