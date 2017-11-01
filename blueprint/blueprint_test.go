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

func TestContainerValueString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		val       interface{}
		expString string
	}{
		{Secret{"foo"}, "Secret: foo"},
		{"bar", "bar"},
		{RuntimeValue{"baz"}, "RuntimeValue: baz"},
	}

	for _, test := range tests {
		res := ContainerValue{test.val}.String()
		assert.Equal(t, test.expString, res)
	}
}

// TestSecretJSON tests marshalling and unmarshalling secrets.
func TestSecretJSON(t *testing.T) {
	t.Parallel()

	secretName := "foo"
	secretJSON := fmt.Sprintf(`{"nameOfSecret": "%s"}`, secretName)

	var unmarshalled ContainerValue
	assert.NoError(t, json.Unmarshal([]byte(secretJSON), &unmarshalled))

	secret, ok := unmarshalled.Value.(Secret)
	assert.True(t, ok)
	assert.Equal(t, secretName, secret.NameOfSecret)
	checkMarshalAndUnmarshal(t, unmarshalled)
}

// TestStringJSON tests marshalling and unmarshalling raw strings.
func TestStringJSON(t *testing.T) {
	t.Parallel()

	str := "bar"
	strJSON := fmt.Sprintf(`"%s"`, str)

	var unmarshalled ContainerValue
	assert.NoError(t, json.Unmarshal([]byte(strJSON), &unmarshalled))

	unmarshalledStr, ok := unmarshalled.Value.(string)
	assert.True(t, ok)
	assert.Equal(t, str, unmarshalledStr)
	checkMarshalAndUnmarshal(t, unmarshalled)
}

// TestRuntimeValueJSON tests marshalling and unmarshalling RuntimeValues.
func TestRuntimeValueJSON(t *testing.T) {
	t.Parallel()

	// Test unmarshalling an invalid resource key.
	var unmarshalled ContainerValue
	err := json.Unmarshal([]byte(`{"resourceKey": "undefined"}`), &unmarshalled)
	assert.Contains(t, err.Error(), "undefined resource key: undefined")

	// Test the success case.
	resourceKey := ContainerPubIPKey
	resourceKeyJSON := fmt.Sprintf(`{"resourceKey": "%s"}`, resourceKey)

	assert.NoError(t, json.Unmarshal([]byte(resourceKeyJSON), &unmarshalled))

	runtimeValue, ok := unmarshalled.Value.(RuntimeValue)
	assert.True(t, ok)
	assert.Equal(t, resourceKey, runtimeValue.ResourceKey)
	checkMarshalAndUnmarshal(t, unmarshalled)
}

// checkMarshalAndUnmarshal checks that that the given ContainerValue marshals
// and unmarshals to the same object.
func checkMarshalAndUnmarshal(t *testing.T, toMarshal ContainerValue) {
	jsonBytes, err := json.Marshal(toMarshal)
	assert.NoError(t, err)

	var unmarshalled ContainerValue
	assert.NoError(t, json.Unmarshal(jsonBytes, &unmarshalled))
	assert.Equal(t, toMarshal, unmarshalled)
}
