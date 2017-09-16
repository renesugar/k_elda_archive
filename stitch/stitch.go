package stitch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

// A Blueprint is an abstract representation of the policy language.
type Blueprint struct {
	Containers    []Container    `json:",omitempty"`
	LoadBalancers []LoadBalancer `json:",omitempty"`
	Connections   []Connection   `json:",omitempty"`
	Placements    []Placement    `json:",omitempty"`
	Machines      []Machine      `json:",omitempty"`

	AdminACL  []string `json:",omitempty"`
	MaxPrice  float64  `json:",omitempty"`
	Namespace string   `json:",omitempty"`
}

// A Placement constraint guides on what type of machine a container can be
// scheduled.
type Placement struct {
	TargetContainerID string `json:",omitempty"`

	Exclusive bool `json:",omitempty"`

	// Machine Constraints
	Provider   string `json:",omitempty"`
	Size       string `json:",omitempty"`
	Region     string `json:",omitempty"`
	FloatingIP string `json:",omitempty"`
}

// An Image represents a Docker image that can be run. If the Dockerfile is non-empty,
// the image should be built and hosted by Quilt.
type Image struct {
	Name       string `json:",omitempty"`
	Dockerfile string `json:",omitempty"`
}

// A Container may be instantiated in the stitch and queried by users.
type Container struct {
	ID                string            `json:",omitempty"`
	Image             Image             `json:",omitempty"`
	Command           []string          `json:",omitempty"`
	Env               map[string]string `json:",omitempty"`
	FilepathToContent map[string]string `json:",omitempty"`
	Hostname          string            `json:",omitempty"`
}

// A LoadBalancer represents a load balanced group of containers.
type LoadBalancer struct {
	Name      string   `json:",omitempty"`
	Hostnames []string `json:",omitempty"`
}

// A Connection allows the container with the `From` hostname to speak to the container
// with the `To` hostname in ports in the range [MinPort, MaxPort]
type Connection struct {
	From    string `json:",omitempty"`
	To      string `json:",omitempty"`
	MinPort int    `json:",omitempty"`
	MaxPort int    `json:",omitempty"`
}

// A ConnectionSlice allows for slices of Collections to be used in joins
type ConnectionSlice []Connection

// A Machine specifies the type of VM that should be booted.
type Machine struct {
	ID          string   `json:",omitempty"`
	Provider    string   `json:",omitempty"`
	Role        string   `json:",omitempty"`
	Size        string   `json:",omitempty"`
	CPU         Range    `json:",omitempty"`
	RAM         Range    `json:",omitempty"`
	DiskSize    int      `json:",omitempty"`
	Region      string   `json:",omitempty"`
	SSHKeys     []string `json:",omitempty"`
	FloatingIP  string   `json:",omitempty"`
	Preemptible bool     `json:",omitempty"`
}

// A Range defines a range of acceptable values for a Machine attribute
type Range struct {
	Min float64 `json:",omitempty"`
	Max float64 `json:",omitempty"`
}

// PublicInternetLabel is a magic label that allows connections to or from the public
// network.
const PublicInternetLabel = "public"

// Accepts returns true if `x` is within the range specified by `stitchr` (include),
// or if no max is specified and `x` is larger than `stitchr.min`.
func (stitchr Range) Accepts(x float64) bool {
	return stitchr.Min <= x && (stitchr.Max == 0 || x <= stitchr.Max)
}

var lookPath = exec.LookPath

// FromFile gets a Blueprint handle from a file on disk.
func FromFile(filename string) (Blueprint, error) {
	if _, err := lookPath("node"); err != nil {
		return Blueprint{}, errors.New(
			"failed to locate Node.js. Is it installed and in your PATH?")
	}

	outFile, err := ioutil.TempFile("", "quilt-out")
	if err != nil {
		return Blueprint{}, fmt.Errorf(
			"failed to create deployment file: %s", err)
	}
	defer outFile.Close()
	defer os.Remove(outFile.Name())

	cmd := exec.Command("node", "-e",
		fmt.Sprintf(
			`require("%s");
			require('fs').writeFileSync("%s",
			  JSON.stringify(global._quiltDeployment.toQuiltRepresentation())
		    );`,
			filename, outFile.Name(),
		),
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return Blueprint{}, err
	}

	depl, err := ioutil.ReadAll(outFile)
	if err != nil {
		return Blueprint{}, fmt.Errorf("failed to read deployment file: %s", err)
	}
	return FromJSON(string(depl))
}

// FromJSON gets a Stitch handle from the deployment representation.
func FromJSON(jsonStr string) (blueprint Blueprint, err error) {
	err = json.Unmarshal([]byte(jsonStr), &blueprint)
	if err != nil {
		err = fmt.Errorf("unable to parse blueprint: %s", err)
	}
	return blueprint, err
}

// String returns the Stitch in its deployment representation.
func (stitch Blueprint) String() string {
	jsonBytes, err := json.Marshal(stitch)
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

// Get returns the value contained at the given index
func (cs ConnectionSlice) Get(ii int) interface{} {
	return cs[ii]
}

// Len returns the number of items in the slice
func (cs ConnectionSlice) Len() int {
	return len(cs)
}
