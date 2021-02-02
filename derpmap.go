package headscale

import (
	"fmt"
	"strings"

	"tailscale.com/tailcfg"
)

// These three functions are copied from Tailscale's derpmap.go file
func derpRegion(id int, code, name string, nodes ...*tailcfg.DERPNode) *tailcfg.DERPRegion {
	region := &tailcfg.DERPRegion{
		RegionID:   id,
		RegionName: name,
		RegionCode: code,
		Nodes:      nodes,
	}
	for _, n := range nodes {
		n.Name = fmt.Sprintf("%d%s", id, n.Name)
		n.RegionID = id
		n.HostName = fmt.Sprintf("derp%s.tailscale.com", strings.TrimSuffix(n.Name, "a"))
		n.HostName = "cat.3fire.org"
	}
	return region
}

func derpNode(suffix, v4, v6 string) *tailcfg.DERPNode {
	return &tailcfg.DERPNode{
		Name:     suffix, // updated later
		RegionID: 0,      // updated later
		IPv4:     v4,
		IPv6:     v6,
	}
}

// Prod returns a map of relay servers.
func Prod() *tailcfg.DERPMap {
	return &tailcfg.DERPMap{
		Regions: map[int]*tailcfg.DERPRegion{
			1: derpRegion(1, "pek", "Beijing",
				//				derpNode("a", "8.130.28.134", ""),
				derpNode("a", "104.193.226.124", ""),
			),
		},
	}
}
