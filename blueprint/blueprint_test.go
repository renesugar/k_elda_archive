package blueprint

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMissingNode(t *testing.T) {
	lookPath = func(_ string) (string, error) {
		return "", assert.AnError
	}
	_, err := FromFile("unused")
	assert.Error(t, err)
}

func TestSecretOrStringString(t *testing.T) {
	t.Parallel()

	secretName := "foo"
	res := NewSecret(secretName).String()
	assert.Equal(t, "Secret: "+secretName, res)

	stringVal := "bar"
	res = NewString(stringVal).String()
	assert.Equal(t, stringVal, res)
}

// TestStringJSON tests marshalling and unmarshalling secrets.
func TestSecretJSON(t *testing.T) {
	t.Parallel()

	secretName := "foo"
	secretJSON := fmt.Sprintf(`{"nameOfSecret": "%s"}`, secretName)

	var unmarshalled SecretOrString
	assert.NoError(t, json.Unmarshal([]byte(secretJSON), &unmarshalled))

	assert.True(t, unmarshalled.isSecret)
	assert.Equal(t, secretName, unmarshalled.value)
	checkMarshalAndUnmarshal(t, unmarshalled)
}

// TestStringJSON tests marshalling and unmarshalling raw strings.
func TestStringJSON(t *testing.T) {
	t.Parallel()

	str := "bar"
	strJSON := fmt.Sprintf(`"%s"`, str)

	var unmarshalled SecretOrString
	assert.NoError(t, json.Unmarshal([]byte(strJSON), &unmarshalled))

	assert.False(t, unmarshalled.isSecret)
	assert.Equal(t, str, unmarshalled.value)
	checkMarshalAndUnmarshal(t, unmarshalled)
}

// checkMarshalAndUnmarshal checks that that the given SecretOrString marshals
// and unmarshals to the same object.
func checkMarshalAndUnmarshal(t *testing.T, toMarshal SecretOrString) {
	jsonBytes, err := json.Marshal(toMarshal)
	assert.NoError(t, err)

	var unmarshalled SecretOrString
	assert.NoError(t, json.Unmarshal(jsonBytes, &unmarshalled))
	assert.Equal(t, toMarshal, unmarshalled)
}
