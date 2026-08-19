package agent

import (
	"net"
	"sort"
)

func AvailableFamilies() ([]int, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var addresses []net.IP
	for _, networkInterface := range interfaces {
		if networkInterface.Flags&net.FlagUp == 0 || networkInterface.Flags&net.FlagLoopback != 0 {
			continue
		}
		assigned, err := networkInterface.Addrs()
		if err != nil {
			continue
		}
		for _, address := range assigned {
			ip, _, err := net.ParseCIDR(address.String())
			if err == nil {
				addresses = append(addresses, ip)
			}
		}
	}
	return familiesFromIPs(addresses), nil
}

func familiesFromIPs(addresses []net.IP) []int {
	seen := map[int]bool{}
	for _, ip := range addresses {
		if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
			continue
		}
		if ip.To4() != nil {
			seen[4] = true
		} else if ip.To16() != nil && ip.IsGlobalUnicast() {
			seen[6] = true
		}
	}
	families := make([]int, 0, 2)
	for family := range seen {
		families = append(families, family)
	}
	sort.Ints(families)
	return families
}
