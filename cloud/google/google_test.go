package google

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kelda/kelda/cloud/acl"
	"github.com/kelda/kelda/cloud/cfg"
	"github.com/kelda/kelda/cloud/google/client/mocks"
	"github.com/kelda/kelda/db"

	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	compute "google.golang.org/api/compute/v1"
)

func getProvider() (*mocks.Client, Provider) {
	backoffWaitFor = func(p func() bool, c time.Duration, t time.Duration) error {
		return nil
	}

	mc := new(mocks.Client)
	return mc, Provider{
		Client:    mc,
		namespace: "namespace",
		network:   "network",
		zone:      "zone-1",
		region:    "region-1",
	}
}

func TestList(t *testing.T) {
	mc, gce := getProvider()
	mc.On("ListInstances", "zone-1", gce.network).Return(&compute.InstanceList{
		Items: []*compute.Instance{
			{
				MachineType: "machine/split/type-1",
				Name:        "name-1",
				NetworkInterfaces: []*compute.NetworkInterface{
					{
						AccessConfigs: []*compute.AccessConfig{
							{
								NatIP: "x.x.x.x",
							},
						},
						NetworkIP: "y.y.y.y",
					},
				},
			},
		},
	}, nil)

	machines, err := gce.List()
	assert.NoError(t, err)
	assert.Len(t, machines, 1)
	assert.Equal(t, machines[0], db.Machine{
		Provider:  "Google",
		Region:    "zone-1",
		CloudID:   "name-1",
		PublicIP:  "x.x.x.x",
		PrivateIP: "y.y.y.y",
		Size:      "type-1",
	})
}

func TestListBadNetworkInterface(t *testing.T) {
	mc, gce := getProvider()

	// Test that List does not return an error (but it logs a warning) when
	// there are no IPs configured.
	mc.On("ListInstances", "zone-1", gce.network).Return(&compute.InstanceList{
		Items: []*compute.Instance{
			{
				MachineType:       "machine/split/type-1",
				Name:              "name-1",
				NetworkInterfaces: []*compute.NetworkInterface{},
			},
		},
	}, nil)
	hook := logrusTest.NewGlobal()

	machines, err := gce.List()
	assert.NoError(t, err)
	assert.Equal(t, []db.Machine{
		{
			Provider: db.Google,
			Region:   "zone-1",
			Size:     "type-1",
			CloudID:  "name-1",
		},
	}, machines)
	assert.Len(t, hook.Entries, 1)
	assert.Equal(t, hook.Entries[0].Level, logrus.WarnLevel)
	assert.Equal(t, hook.Entries[0].Message, "Failed to get machine IP")
}

func TestParseACL(t *testing.T) {
	_, gce := getProvider()

	_, err := gce.parseACL(&compute.Firewall{})
	assert.EqualError(t, err, "malformed firewall")

	// Missing ports.
	_, err = gce.parseACL(&compute.Firewall{
		SourceRanges: []string{"1.2.3.4/32"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "icmp",
		}, {
			IPProtocol: "udp",
		}, {
			IPProtocol: "tcp",
		}},
	})
	assert.EqualError(t, err, "malformed firewall")

	// Unequal ports.
	_, err = gce.parseACL(&compute.Firewall{
		SourceRanges: []string{"1.2.3.4/32"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "icmp",
		}, {
			IPProtocol: "udp",
			Ports:      []string{"80"},
		}, {
			IPProtocol: "tcp",
			Ports:      []string{"90"},
		}},
	})
	assert.EqualError(t, err, "malformed firewall")

	// Bad Ports
	_, err = gce.parseACL(&compute.Firewall{
		SourceRanges: []string{"1.2.3.4/32"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "icmp",
		}, {
			IPProtocol: "udp",
			Ports:      []string{"rabbit"},
		}, {
			IPProtocol: "tcp",
			Ports:      []string{"rabbit"},
		}},
	})
	assert.EqualError(t, err, "invalid port: rabbit")

	// Bad dashed port
	_, err = gce.parseACL(&compute.Firewall{
		SourceRanges: []string{"1.2.3.4/32"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "icmp",
		}, {
			IPProtocol: "udp",
			Ports:      []string{"1-2-3"},
		}, {
			IPProtocol: "tcp",
			Ports:      []string{"1-2-3"},
		}},
	})
	assert.EqualError(t, err, "invalid port: 1-2-3")

	// Single Port
	gacl, err := gce.parseACL(&compute.Firewall{
		Name:         "name",
		SourceRanges: []string{"1.2.3.4/32"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "icmp",
		}, {
			IPProtocol: "udp",
			Ports:      []string{"1"},
		}, {
			IPProtocol: "tcp",
			Ports:      []string{"1"},
		}},
	})
	assert.NoError(t, err)
	assert.Equal(t, gACL{"name", acl.ACL{CidrIP: "1.2.3.4/32",
		MinPort: 1, MaxPort: 1}}, gacl)

	// Double Port
	gacl, err = gce.parseACL(&compute.Firewall{
		Name:         "name",
		SourceRanges: []string{"1.2.3.4/32"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "icmp",
		}, {
			IPProtocol: "udp",
			Ports:      []string{"1-2"},
		}, {
			IPProtocol: "tcp",
			Ports:      []string{"1-2"},
		}},
	})
	assert.NoError(t, err)
	assert.Equal(t, gACL{"name", acl.ACL{CidrIP: "1.2.3.4/32",
		MinPort: 1, MaxPort: 2}}, gacl)
}

func TestSetACLs(t *testing.T) {
	mc, gce := getProvider()

	mc.On("ListFirewalls", mock.Anything).Return(nil, errors.New("err")).Once()
	err := gce.setACLs(nil)
	assert.EqualError(t, err, "list firewalls: err")

	mc.On("ListFirewalls", gce.network).Return(&compute.FirewallList{
		Items: []*compute.Firewall{{Name: "Delete"}},
	}, nil)

	mc.On("DeleteFirewall", "Delete").Return(nil, errors.New("delete")).Once()
	err = gce.setACLs(nil)
	assert.EqualError(t, err, "delete")

	mc.On("DeleteFirewall", "Delete").Return(nil, nil)

	mc.On("InsertFirewall", mock.Anything).Return(nil, errors.New("insert")).Once()
	err = gce.setACLs([]acl.ACL{{CidrIP: "1.2.3.4/32"}})
	assert.EqualError(t, err, "insert")

	mc.On("InsertFirewall", &compute.Firewall{
		Name:         "network-1-2-3-4-32-0-0",
		Network:      gce.networkURL(),
		Description:  gce.network,
		SourceRanges: []string{"1.2.3.4/32"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"0-0"},
		}, {
			IPProtocol: "udp",
			Ports:      []string{"0-0"},
		}, {
			IPProtocol: "icmp",
		}}}).Return(nil, nil)
	err = gce.setACLs([]acl.ACL{{CidrIP: "1.2.3.4/32"}})
	assert.NoError(t, err)

	// Verify internal firewall rule gets installed.
	mc.On("InsertFirewall", &compute.Firewall{
		Name:         "network-172-16-0-0-12-0-65535",
		Network:      gce.networkURL(),
		Description:  gce.network,
		SourceRanges: []string{ipv4Range},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"0-65535"},
		}, {
			IPProtocol: "udp",
			Ports:      []string{"0-65535"},
		}, {
			IPProtocol: "icmp",
		}}}).Return(nil, nil)
	err = gce.SetACLs(nil)
	assert.NoError(t, err)
	mc.AssertExpectations(t)
}

func TestPlanSetACLs(t *testing.T) {
	_, gce := getProvider()
	add, remove := gce.planSetACLs([]*compute.Firewall{{
		Name: "Unparseable",
	}, {
		Name:         "Delete",
		SourceRanges: []string{"1.2.3.4/32"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "icmp",
		}, {
			IPProtocol: "udp",
			Ports:      []string{"1"},
		}, {
			IPProtocol: "tcp",
			Ports:      []string{"1"},
		}},
	}, {
		Name:         "network-5-6-7-8-32-1-2",
		SourceRanges: []string{"5.6.7.8/32"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "icmp",
		}, {
			IPProtocol: "udp",
			Ports:      []string{"1-2"},
		}, {
			IPProtocol: "tcp",
			Ports:      []string{"1-2"},
		}},
	}}, []acl.ACL{{
		CidrIP:  "5.6.7.8/32",
		MinPort: 1,
		MaxPort: 2,
	}, {
		CidrIP:  "9.9.9.9/32",
		MinPort: 3,
		MaxPort: 4,
	}})
	assert.Equal(t, []string{"Unparseable", "Delete"}, remove)
	assert.Equal(t, []*compute.Firewall{{
		Name:         "network-9-9-9-9-32-3-4",
		Network:      gce.networkURL(),
		Description:  gce.network,
		SourceRanges: []string{"9.9.9.9/32"},
		Allowed: []*compute.FirewallAllowed{{
			IPProtocol: "tcp",
			Ports:      []string{"3-4"},
		}, {
			IPProtocol: "udp",
			Ports:      []string{"3-4"},
		}, {
			IPProtocol: "icmp",
		}},
	}}, add)
}

func TestBoot(t *testing.T) {
	mc, gce := getProvider()

	mc.On("ListNetworks", mock.Anything).Return(nil, errors.New("list err")).Once()
	_, err := gce.Boot(nil)
	assert.EqualError(t, err, "list err")

	mc.On("ListNetworks", mock.Anything).Return(&compute.NetworkList{
		Items: []*compute.Network{{Name: gce.network}},
	}, nil)

	_, err = gce.Boot([]db.Machine{{Preemptible: true}})
	assert.EqualError(t, err, "preemptible vms are not implemented")

	mc.On("InsertInstance", "zone-1", mock.Anything).Return(
		nil, errors.New("err")).Once()

	_, err = gce.Boot([]db.Machine{{Size: "size1"}})
	assert.EqualError(t, err, "err")

	name := 0
	randName = func() string {
		name++
		return fmt.Sprintf("%d", name)
	}

	machines := []db.Machine{{Size: "size1"}, {Size: "size2"}}

	cfg1 := gce.instanceConfig("1", "size1", cfg.Ubuntu(machines[0], ""))
	mc.On("InsertInstance", "zone-1", cfg1).Return(nil, nil)

	cfg2 := gce.instanceConfig("2", "size2", cfg.Ubuntu(machines[1], ""))
	mc.On("InsertInstance", "zone-1", cfg2).Return(nil, nil)

	ids, err := gce.Boot(machines)
	assert.NoError(t, err)
	assert.Equal(t, []string{"1", "2"}, ids)

	mc.AssertExpectations(t)
}

func TestStop(t *testing.T) {
	mc, gce := getProvider()

	mc.On("DeleteInstance", mock.Anything, mock.Anything).Return(
		nil, errors.New("err")).Once()
	err := gce.Stop([]db.Machine{{CloudID: "1"}})
	assert.EqualError(t, err, "err")

	mc.On("DeleteInstance", "zone-1", "1").Return(nil, nil)
	mc.On("DeleteInstance", "zone-1", "2").Return(nil, nil)
	assert.NoError(t, gce.Stop([]db.Machine{{CloudID: "1"}, {CloudID: "2"}}))
	mc.AssertExpectations(t)
}

func TestInstanceConfig(t *testing.T) {
	_, gce := getProvider()
	cloudConfig := "cloudConfig"
	res := gce.instanceConfig("name", "size", cloudConfig)
	exp := &compute.Instance{
		Name:        "name",
		Description: gce.network,
		MachineType: "zones/zone-1/machineTypes/size",
		Disks: []*compute.AttachedDisk{{
			Boot:       true,
			AutoDelete: true,
			InitializeParams: &compute.AttachedDiskInitializeParams{
				SourceImage: image,
			},
		}},
		NetworkInterfaces: []*compute.NetworkInterface{{
			AccessConfigs: []*compute.AccessConfig{{
				Type: "ONE_TO_ONE_NAT",
				Name: ephemeralIPName,
			}},
			Network: gce.networkURL(),
		}},
		Metadata: &compute.Metadata{
			Items: []*compute.MetadataItems{{
				Key:   "startup-script",
				Value: &cloudConfig,
			}},
		},
	}

	assert.Equal(t, exp, res)
}

func TestCleanup(t *testing.T) {
	mc, gce := getProvider()

	mc.On("ListNetworks", gce.network).Return(nil, errors.New("err")).Once()
	assert.EqualError(t, gce.Cleanup(), "err")

	// Do nothing if there are no networks listed.
	mc.On("ListNetworks", gce.network).Return(&compute.NetworkList{
		Items: []*compute.Network{}}, nil).Once()
	assert.NoError(t, gce.Cleanup())

	mc.On("ListNetworks", gce.network).Return(&compute.NetworkList{
		Items: []*compute.Network{{Name: gce.network}}}, nil)

	mc.On("ListFirewalls", gce.network).Return(nil, errors.New("lf")).Once()
	assert.EqualError(t, gce.Cleanup(), "list firewalls: lf")

	mc.On("ListFirewalls", gce.network).Return(&compute.FirewallList{}, nil)
	mc.On("DeleteNetwork", gce.network).Return(nil, errors.New("del")).Once()
	assert.EqualError(t, gce.Cleanup(), "del")

	mc.On("DeleteNetwork", gce.network).Return(nil, nil)
	assert.NoError(t, gce.Cleanup())
}

func TestCreateNetwork(t *testing.T) {
	mc, gce := getProvider()

	mc.On("ListNetworks", gce.network).Return(nil, errors.New("err")).Once()
	assert.EqualError(t, gce.createNetwork(), "err")

	// Do nothing if the network already exists
	mc.On("ListNetworks", gce.network).Return(&compute.NetworkList{
		Items: []*compute.Network{{Name: gce.network}}}, nil).Once()
	assert.NoError(t, gce.createNetwork())

	mc.On("ListNetworks", gce.network).Return(&compute.NetworkList{
		Items: []*compute.Network{}}, nil)
	mc.On("InsertNetwork", mock.Anything).Return(nil, errors.New("err")).Once()
	assert.EqualError(t, gce.createNetwork(), "err")

	mc.On("InsertNetwork", &compute.Network{
		Name: gce.network, IPv4Range: ipv4Range}).Return(nil, nil)
	assert.NoError(t, gce.createNetwork())
	mc.AssertExpectations(t)
}

func TestUpdateFloatingIP(t *testing.T) {
	mc, gce := getProvider()

	desiredIP := "8.8.8.8"
	// Test that attempting to assign an IP that is not reserved by the user
	// throws an error.
	mc.On("ListFloatingIPs", gce.region).Return(&compute.AddressList{}, nil).Once()
	err := gce.UpdateFloatingIPs([]db.Machine{{FloatingIP: desiredIP}})
	assert.EqualError(t, err, "8.8.8.8 is not reserved")

	// Test that attempting to assign an IP that is already assigned throws an
	// error.
	mc.On("ListFloatingIPs", gce.region).Return(&compute.AddressList{
		Items: []*compute.Address{
			{Address: desiredIP, Status: "ASSIGNED"},
		},
	}, nil).Once()
	err = gce.UpdateFloatingIPs([]db.Machine{{FloatingIP: desiredIP}})
	assert.EqualError(t, err, "8.8.8.8 is already assigned")

	// Test that an IP in another invalid state throws an error.
	mc.On("ListFloatingIPs", gce.region).Return(&compute.AddressList{
		Items: []*compute.Address{
			{Address: desiredIP, Status: "PENDING"},
		},
	}, nil).Once()
	err = gce.UpdateFloatingIPs([]db.Machine{{FloatingIP: desiredIP}})
	assert.EqualError(t, err, "8.8.8.8 is not ready to be assigned "+
		"(current IP status: PENDING)")

	// Test the success case.
	cloudID := "cloudID"
	networkIntfName := "eth0"
	mc.On("ListFloatingIPs", gce.region).Return(&compute.AddressList{
		Items: []*compute.Address{
			{Address: desiredIP, Status: "RESERVED"},
		},
	}, nil).Once()
	mc.On("GetInstance", gce.zone, cloudID).Return(&compute.Instance{
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				Name: networkIntfName,
				AccessConfigs: []*compute.AccessConfig{
					{Name: ephemeralIPName},
				},
			},
		},
	}, nil).Once()

	// The old access config should be removed, and an access config for the
	// new IP should be added.
	expAccessConfig := &compute.AccessConfig{
		Name:  floatingIPName,
		Type:  "ONE_TO_ONE_NAT",
		NatIP: desiredIP,
	}
	mc.On("AddAccessConfig", gce.zone, cloudID, networkIntfName,
		expAccessConfig).Return(&compute.Operation{Zone: gce.zone}, nil).Once()
	mc.On("DeleteAccessConfig", gce.zone, cloudID, ephemeralIPName,
		networkIntfName).Return(&compute.Operation{Zone: gce.zone}, nil).Once()

	err = gce.UpdateFloatingIPs([]db.Machine{
		{CloudID: cloudID, FloatingIP: desiredIP}})
	assert.NoError(t, err)
	mc.AssertExpectations(t)
}
