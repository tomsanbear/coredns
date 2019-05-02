package bind

import (
	"fmt"
	"net"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/mholt/caddy"
)

func setup(c *caddy.Controller) error {
	config := dnsserver.GetConfig(c)

	// addresses will be consolidated over all BIND directives available in that BlocServer
	bindable := []string{}
	for c.Next() {
		args := c.RemainingArgs()

		if len(args) == 0 {
			return plugin.Error("bind", fmt.Errorf("at least one address is expected"))
		}
		for _, arg := range args {
			// check if user specified
			if iface, _ := net.InterfaceByName(arg); iface != nil {
				addrs, err := iface.Addrs()
				if err != nil {
					return plugin.Error("bind", fmt.Errorf("no addresses found on interface %v", iface.Name))
				}
				var validAddrs []string
				for _, addr := range addrs {

					validAddr := isValidAddress(addr)
					if validAddr != "" {
						validAddrs = append(validAddrs, validAddr)
					}
				}
				if len(validAddrs) < 1 {
					return plugin.Error("bind", fmt.Errorf("addresses on interface %v were found but not valid: %v", iface.Name, addrs))
				}
				bindable = append(bindable, validAddrs...)
				continue
			} else if net.ParseIP(arg) == nil {
				return plugin.Error("bind", fmt.Errorf("not a valid IP address: %s", arg))
			}
			bindable = append(bindable, arg)
		}
	}
	config.ListenHosts = bindable
	return nil
}

func isValidAddress(address net.Addr) string {
	// Parse out the IP from the address
	ip, _, err := net.ParseCIDR(address.String())
	if err != nil {
		return ""
	}
	return ip.String()
}
