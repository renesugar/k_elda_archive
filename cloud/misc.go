package cloud

import (
	"fmt"

	"github.com/quilt/quilt/cloud/amazon"
	"github.com/quilt/quilt/cloud/digitalocean"
	"github.com/quilt/quilt/cloud/google"
	"github.com/quilt/quilt/cloud/machine"
	"github.com/quilt/quilt/db"
)

// DefaultRegion populates `m.Region` for the provided db.Machine if one isn't
// specified. This is intended to allow users to omit the cloud provider region when
// they don't particularly care where a system is placed.
func DefaultRegion(m db.Machine) db.Machine {
	if m.Region != "" {
		return m
	}

	switch m.Provider {
	case db.Amazon:
		m.Region = amazon.DefaultRegion
	case db.DigitalOcean:
		m.Region = digitalocean.DefaultRegion
	case db.Google:
		m.Region = google.DefaultRegion
	case db.Vagrant:
	default:
		panic(fmt.Sprintf("Unknown Cloud Provider: %s", m.Provider))
	}

	return m
}

// ChooseSize returns an acceptable machine size for the given provider that fits the
// provided ram, cpu, and price constraints.
var ChooseSize = machine.ChooseSize
