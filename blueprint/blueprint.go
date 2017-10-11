package blueprint

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
// the image should be built and hosted by Kelda.
type Image struct {
	Name       string `json:",omitempty"`
	Dockerfile string `json:",omitempty"`
}

// A Container may be instantiated in the blueprint and queried by users.
type Container struct {
	ID                string                    `json:",omitempty"`
	Image             Image                     `json:",omitempty"`
	Command           []string                  `json:",omitempty"`
	Env               map[string]SecretOrString `json:",omitempty"`
	FilepathToContent map[string]SecretOrString `json:",omitempty"`
	Hostname          string                    `json:",omitempty"`
}

// NewSecret returns a SecretOrString representing a secret.
func NewSecret(name string) SecretOrString {
	return SecretOrString{
		value:    name,
		isSecret: true,
	}
}

// NewString returns a SecretOrString representing a string.
func NewString(str string) SecretOrString {
	return SecretOrString{value: str}
}

// SecretOrString represents a value that might be a string, or a Secret.
type SecretOrString struct {
	// If the value refers to a secret, then the value is the name of the
	// secret, and the caller is expected to query Vault to resolve the secret
	// value.
	value    string
	isSecret bool
}

// Value returns the value of the SecretOrString, and whether the value refers
// to a secret. If the value is a secret, then the name of the secret is
// returned, and the caller is expected to query Vault to resolve the secret
// value. Otherwise, the value is used directly.
func (secretOrString SecretOrString) Value() (string, bool) {
	return secretOrString.value, secretOrString.isSecret
}

// String returns a human-readable representation of the secretOrString. This
// makes the database logs easier to read.
func (secretOrString SecretOrString) String() string {
	val, isSecret := secretOrString.Value()
	if isSecret {
		return "Secret: " + val
	}
	return val
}

// secret is used to marshal and unmarshal secrets between the JavaScript
// API and the deployment engine. The name of the field must match the name
// in the JavaScript Secret object.
type secret struct {
	NameOfSecret string
}

// UnmarshalJSON implements the JSON unmarshal interface. It first attempts
// to convert the JSON object into a secret. If that fails, it tries
// to convert the object into a string.
func (secretOrString *SecretOrString) UnmarshalJSON(jsonBytes []byte) error {
	trySecret := secret{}
	secretErr := json.Unmarshal(jsonBytes, &trySecret)
	if secretErr == nil {
		secretOrString.isSecret = true
		secretOrString.value = trySecret.NameOfSecret
		return nil
	}

	var tryString string
	stringErr := json.Unmarshal(jsonBytes, &tryString)
	if stringErr == nil {
		secretOrString.value = tryString
		return nil
	}

	return fmt.Errorf("neither secret (%s), nor string (%s)", secretErr, stringErr)
}

// MarshalJSON implements the JSON marshal interface. If the value is a secret,
// it converts the value into a secret object. Otherwise, it sends the raw
// string directly.
func (secretOrString SecretOrString) MarshalJSON() ([]byte, error) {
	if secretOrString.isSecret {
		return json.Marshal(secret{secretOrString.value})
	}
	return json.Marshal(secretOrString.value)
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
	DiskSize    int      `json:",omitempty"`
	Region      string   `json:",omitempty"`
	SSHKeys     []string `json:",omitempty"`
	FloatingIP  string   `json:",omitempty"`
	Preemptible bool     `json:",omitempty"`
}

// PublicInternetLabel is a magic label that allows connections to or from the public
// network.
const PublicInternetLabel = "public"

var lookPath = exec.LookPath

// FromFile gets a Blueprint handle from a file on disk.
func FromFile(filename string) (Blueprint, error) {
	if _, err := lookPath("node"); err != nil {
		return Blueprint{}, errors.New(
			"failed to locate Node.js. Is it installed and in your PATH?")
	}

	outFile, err := ioutil.TempFile("", "kelda-out")
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
			  JSON.stringify(global._keldaDeployment.toKeldaRepresentation())
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

// FromJSON gets a Blueprint handle from the deployment representation.
func FromJSON(jsonStr string) (bp Blueprint, err error) {
	err = json.Unmarshal([]byte(jsonStr), &bp)
	if err != nil {
		err = fmt.Errorf("unable to parse blueprint: %s", err)
	}
	return bp, err
}

// String returns the Blueprint in its deployment representation.
func (bp Blueprint) String() string {
	jsonBytes, err := json.Marshal(bp)
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
