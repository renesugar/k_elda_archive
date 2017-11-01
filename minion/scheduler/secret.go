package scheduler

import (
	"time"

	"github.com/kelda/kelda/blueprint"
	"github.com/kelda/kelda/db"
	"github.com/kelda/kelda/minion/vault"

	log "github.com/sirupsen/logrus"
)

var secretCache = map[string]cacheEntry{}

const cacheTimeout = 1 * time.Minute

type cacheEntry struct {
	value      string
	expiration time.Time
}

func newCacheEntry(value string) cacheEntry {
	return cacheEntry{value, time.Now().Add(cacheTimeout)}
}

func (entry cacheEntry) isValid() bool {
	return time.Now().Before(entry.expiration)
}

// resolveSecrets attempts to create a map of all secret names to values needed
// for the provided containers to be run.
func resolveSecrets(client vault.SecretStore, dbcs []db.Container) map[string]string {
	secretMap := map[string]string{}
	for _, dbc := range dbcs {
		for _, name := range dbc.GetReferencedSecrets() {
			if _, ok := secretMap[name]; ok {
				continue
			}

			secretVal, err := getSecret(client, name)
			if err == nil {
				secretMap[name] = secretVal
				continue
			}

			// It is expected for secrets to not exist in Vault before users
			// run the `kelda secret` command. Therefore, we only log the
			// error if it is for a reason other than not running `kelda
			// secret`.
			if err == vault.ErrSecretDoesNotExist {
				continue
			}

			log.WithFields(log.Fields{
				"error": err,
				"name":  name,
			}).Info("Failed to fetch secret. This error is probably benign " +
				"if the container was launched recently -- permission " +
				"issues are expected when containers are first " +
				"scheduled because the leader is still configuring the " +
				"Vault ACLs for accessing the secret.")
		}
	}
	return secretMap
}

// getSecret returns the value of the associated with `name`. If there is a
// non-expired cache entry for `name` from a previous successful `getSecret`
// call, the cached version is returned.
func getSecret(secretStore vault.SecretStore, name string) (string, error) {
	if cacheEntry, ok := secretCache[name]; ok && cacheEntry.isValid() {
		return cacheEntry.value, nil
	}

	secret, err := secretStore.Read(name)
	if err != nil {
		return "", err
	}

	secretCache[name] = newCacheEntry(secret)
	return secret, nil
}

// evaluateContainerValues converts a map with ContainerValue values into raw
// strings. It does so by looking up the value of secrets in the given secretMap,
// and converting any RuntimeValue values.
// Any undefined secrets are returned in the `missing` slice.
func evaluateContainerValues(toEvaluate map[string]blueprint.ContainerValue,
	secretMap map[string]string, myPubIP string) (map[string]string, []string) {

	var missing []string
	resolved := map[string]string{}
	for key, valIntf := range toEvaluate {
		switch val := valIntf.Value.(type) {
		case blueprint.Secret:
			secret, ok := secretMap[val.NameOfSecret]
			if !ok {
				missing = append(missing, val.NameOfSecret)
				continue
			}
			resolved[key] = secret
		case blueprint.RuntimeValue:
			if val.ResourceKey == blueprint.ContainerPubIPKey {
				resolved[key] = myPubIP
			} else {
				log.WithField("key", val.ResourceKey).
					Warn("Unknown RuntimeValue key")
			}
		case string:
			resolved[key] = val
		default:
			panic("unexpected container value type")
		}
	}
	return resolved, missing
}
