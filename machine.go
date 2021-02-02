package headscale

/*
The IPN network contains 3 concepts: Machine, Node and User.
One Machine maps to one physical device
One Node maps to one logical WireGuard interface
One User maps to one user account in the system, it can have multiple Logins
such as DingTalk, WeChat, uid/password

Both User and Node data structure are defined in tailcfg module

A Machine can contain multiple Nodes. Each Node belongs to one and only one User.
Each User can have one or more Nodes reside on different Machine.

Login(Register) process: endpoint:Machine Key, Bind Machine and Node，Authorize
Node/Machine if needed. If found binding doesn't match or Authinfo expires, then
ask User to Authenticate.
Logout process: Unbind Machine and Node. Client delete state file, next time a
new Node Key will be generated and need to bind to Machine again through Login.
Map Process: endpoint: Machine Key, Index Key: NodeKey. Get back list of peers.
This process is the most relavent part to IAM.
*/

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	"inet.af/netaddr"
	"tailscale.com/tailcfg"
	"tailscale.com/wgengine/wgcfg"
)

// Machine struct represent a local machine identified by MachineKey
type Machine struct {
	ID         uint64 `gorm:"primary_key"`
	MachineKey string `gorm:"type:varchar(64);unique_index"`
	NodeKey    string
	DiscoKey   string
	IPAddress  string

	Registered bool // temp
	LastSeen   *time.Time
	Expiry     *time.Time

	HostInfo  postgres.Jsonb
	Endpoints postgres.Jsonb

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// For the time being this method is rather naive
func (m Machine) isAlreadyRegistered() bool {
	return m.Registered
}

func (m Machine) toNode() (*tailcfg.Node, error) {
	nKey, err := wgcfg.ParseHexKey(m.NodeKey)
	if err != nil {
		return nil, err
	}
	mKey, err := wgcfg.ParseHexKey(m.MachineKey)
	if err != nil {
		return nil, err
	}
	dKey, err := wgcfg.ParseHexKey(m.DiscoKey)
	if err != nil {
		return nil, err
	}
	addrs := []netaddr.IPPrefix{}
	allowedIPs := []netaddr.IPPrefix{}

	ip, err := netaddr.ParseIPPrefix(fmt.Sprintf("%s/32", m.IPAddress))
	if err != nil {
		return nil, err
	}
	addrs = append(addrs, ip)
	allowedIPs = append(allowedIPs, ip) // looks like the client expect this

	endpoints := []string{}
	be, err := m.Endpoints.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(be, &endpoints)
	if err != nil {
		return nil, err
	}

	hostinfo := tailcfg.Hostinfo{}
	hi, err := m.HostInfo.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(hi, &hostinfo)
	if err != nil {
		return nil, err
	}

	n := tailcfg.Node{
		ID:         tailcfg.NodeID(m.ID), // this is the actual ID
		Name:       "",
		User:       1,
		Key:        tailcfg.NodeKey(nKey),
		KeyExpiry:  *m.Expiry,
		Machine:    tailcfg.MachineKey(mKey),
		DiscoKey:   tailcfg.DiscoKey(dKey),
		Addresses:  addrs,
		AllowedIPs: allowedIPs,
		Endpoints:  endpoints,
		DERP:       "127.3.3.40:1", // wtf?

		Hostinfo: hostinfo,
		Created:  m.CreatedAt,
		LastSeen: m.LastSeen,

		KeepAlive:         false,
		MachineAuthorized: m.Registered,
	}
	// n.Key.MarshalText()
	return &n, nil
}

func (h *Headscale) getPeers(m Machine) (*[]*tailcfg.Node, error) {
	db, err := h.db()
	if err != nil {
		log.Printf("Cannot open DB: %s", err)
		return nil, err
	}
	defer db.Close()

	// Add user management here?
	machines := []Machine{}
	if err = db.Where("machine_key <> ? AND registered", m.MachineKey).Find(&machines).Error; err != nil {
		log.Printf("Error accessing db: %s", err)
		return nil, err
	}

	log.Printf("Found %d peers of %s", len(machines), m.MachineKey)

	peers := []*tailcfg.Node{}
	for _, mn := range machines {
		peer, err := mn.toNode()
		if err != nil {
			return nil, err
		}
		peers = append(peers, peer)
	}
	sort.Slice(peers, func(i, j int) bool { return peers[i].ID < peers[j].ID })
	return &peers, nil
}
