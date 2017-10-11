package vault

import (
	"fmt"
	"path"
	"strings"

	"github.com/kelda/kelda/db"
	"github.com/kelda/kelda/join"

	log "github.com/sirupsen/logrus"
)

var (
	// Vault locations used for authentication. An authentication backend runs
	// in the Vault container, and stores authentication information in Vault
	// under the path /auth/<certMountName>. certMountName is something we can
	// choose: the other three paths are dictated by Vault (for example, if we
	// set certMountName to `foo`, the login information would be stored at
	// `/auth/foo/login`).
	certMountName     = "cert"
	certRootPath      = path.Join("/auth", certMountName)
	certLoginEndpoint = path.Join(certRootPath, "login")
	certListEndpoint  = path.Join(certRootPath, "certs")
)

// certRole represents the access configuration for an entity with the private
// key associated with the cert.
type certRole struct {
	// The identifier for the certificate role. This does not affect the behavior
	// of the role -- it is simply for human readability.
	name string

	// The public key associated with the role. Clients authenticate using the
	// private key associate with this public key.
	cert string

	// The policies associated with the client once it authenticates. The
	// policies define the actual ACL restrictions. See policy.go for more
	// details.
	policies []string
}

func syncAuth(vaultClient APIClient, conn db.Conn) {
	currentRoles, err := getCurrentRoles(vaultClient)
	if err != nil {
		log.WithError(err).Error("Failed to get current Vault roles")
		return
	}

	joinRoles(vaultClient, getDesiredRoles(conn), currentRoles)
}

func getCurrentRoles(vaultClient APIClient) ([]certRole, error) {
	listAuthResp, err := vaultClient.List(certListEndpoint)
	if err != nil {
		return nil, err
	}

	// If no roles have been configured, the list will return nil.
	if listAuthResp == nil {
		return nil, nil
	}

	var roles []certRole
	for _, roleNameIntf := range listAuthResp.Data["keys"].([]interface{}) {
		roleName := roleNameIntf.(string)
		role, err := vaultClient.Read(pathForRole(roleName))
		if err != nil {
			return nil, err
		}

		var policies []string
		for _, policyIntf := range role.Data["policies"].([]interface{}) {
			policies = append(policies, policyIntf.(string))
		}
		roles = append(roles, certRole{
			name:     roleName,
			cert:     role.Data["certificate"].(string),
			policies: policies,
		})
	}
	return roles, nil
}

func getDesiredRoles(conn db.Conn) (desiredRoles []certRole) {
	// Associate each minion's public key with the policy in its name.
	for minion, cert := range conn.MinionSelf().MinionIPToPublicKey {
		desiredRoles = append(desiredRoles, certRole{
			name:     minion,
			cert:     cert,
			policies: []string{minion},
		})
	}
	return desiredRoles
}

func joinRoles(vaultClient APIClient, desiredRoles, currentRoles certRoleSlice) {
	key := func(roleIntf interface{}) interface{} {
		role := roleIntf.(certRole)
		return struct{ name, cert, policies string }{
			role.name, role.cert, fmt.Sprintf("%+v", role.policies),
		}
	}
	_, toAdd, toDelete := join.HashJoin(desiredRoles, currentRoles,
		key, key)

	for _, roleIntf := range toAdd {
		role := roleIntf.(certRole)
		_, err := vaultClient.Write(pathForRole(role.name),
			map[string]interface{}{
				"certificate": role.cert,
				"policies":    strings.Join(role.policies, ","),
			})
		if err != nil {
			log.WithError(err).WithField("role", role).Error(
				"Failed to create Vault authentication role")
		}
	}

	for _, role := range toDelete {
		roleName := role.(certRole).name
		_, err := vaultClient.Delete(pathForRole(roleName))
		if err != nil {
			log.WithError(err).WithField("role", role).Error(
				"Failed to delete Vault authentication role")
		}
	}
}

func pathForRole(role string) string {
	return path.Join(certListEndpoint, role)
}

type certRoleSlice []certRole

// Get returns the value contained at the given index
func (cs certRoleSlice) Get(i int) interface{} {
	return cs[i]
}

// Len returns the number of items in the slice.
func (cs certRoleSlice) Len() int {
	return len(cs)
}
