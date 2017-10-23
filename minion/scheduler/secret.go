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

// evaluateSecretOrStrings converts a map with SecretOrString values into raw
// strings. It does so by looking up the value of secrets in the given secretMap.
// Any undefined secrets are returned in the `missing` slice.
func evaluateSecretOrStrings(toEvaluate map[string]blueprint.SecretOrString,
	secretMap map[string]string) (resolved map[string]string, missing []string) {

	resolved = map[string]string{}
	for key, val := range toEvaluate {
		secretOrString, isSecret := val.Value()

		// If boxedVal is a secret, then we need to look it up in the secretMap.
		// Otherwise, the value is a raw string, and we can use it directly.
		if isSecret {
			secret, ok := secretMap[secretOrString]
			if !ok {
				missing = append(missing, secretOrString)
				continue
			}
			resolved[key] = secret
		} else {
			resolved[key] = secretOrString
		}
	}
	return resolved, missing
}
