package headscale

import (
	"fmt"
	"strings"

	"tailscale.com/tailcfg"
)

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

// Prod returns Tailscale's map of relay servers.
//
// This list is only used by cmd/tailscale's netcheck subcommand. In
// normal operation the Tailscale nodes get this sent to them from the
// control server.
//
// This list is subject to change and should not be relied on.
func Prod() *tailcfg.DERPMap {
	return &tailcfg.DERPMap{
		Regions: map[int]*tailcfg.DERPRegion{
			1: derpRegion(1, "pek", "Beijing",
				derpNode("a", "8.130.28.134", ""),
			),
		},
	}
}
