package plugin

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator_LoadSchema(t *testing.T) {
	v, err := NewValidator()
	assert.Nil(t, err)

	t.Run("loads a good schema", func(t *testing.T) {
		schema, err := ioutil.ReadFile("testdata/good_schema.lua")
		assert.Nil(t, err)
		assert.Nil(t, v.LoadSchema(string(schema)))
	})
	t.Run("fails to load a bad schema", func(t *testing.T) {
		schema, err := ioutil.ReadFile("testdata/bad_schema.lua")
		assert.Nil(t, err)
		err = v.LoadSchema(string(schema))
		assert.NotNil(t, err)
		expected := "name: field required for entity check"
		assert.True(t, strings.Contains(err.Error(), expected))
	})
}

func TestValidator_Validate(t *testing.T) {
	v, err := NewValidator()
	assert.Nil(t, err)
	schema, err := ioutil.ReadFile("testdata/uuid_schema.lua")
	assert.Nil(t, err)
	assert.Nil(t, v.LoadSchema(string(schema)))
	t.Run("validates a uuid correctly", func(t *testing.T) {
		plugin := `{
                     "name": "test",
                     "config": {
                       "uuid": "bar"
                     },
                     "enabled": true,
                     "protocols": [ "http"]
                   }`
		err := v.Validate(plugin)
		uuidErr := getErrForField(err, "config.uuid")
		assert.Equal(t, "expected a valid UUID", uuidErr)
	})
	t.Run("no error for a correct uuid", func(t *testing.T) {
		plugin := `{
                     "name": "test",
                     "config": {
                       "uuid": "ad7d1dcd-9511-4310-9803-356279a7ae32"
                     },
                     "enabled": true,
                     "protocols": [ "http"]
                   }`
		err := v.Validate(plugin)
		assert.Nil(t, err)
	})
}

func TestValidator_ProcessAutoFields(t *testing.T) {
	v, err := NewValidator()
	assert.Nil(t, err)
	schema, err := ioutil.ReadFile("testdata/good_schema.lua")
	assert.Nil(t, err)
	assert.Nil(t, v.LoadSchema(string(schema)))
	t.Run("populates defaults for key-auth plugin", func(t *testing.T) {
		plugin := `{
                     "name": "key-auth",
                     "config": {}
                   }`
		plugin, err := v.ProcessAutoFields(plugin)
		assert.Nil(t, err)
		expected := `{"config":{"hide_credentials":false,"key_in_body":false,` +
			`"key_in_header":true,"key_in_query":true,"key_names":["apikey"],` +
			`"run_on_preflight":true},"enabled":true,"name":"key-auth",` +
			`"protocols":["grpc","grpcs","http","https"]}`
		assert.Equal(t, expected, plugin)
	})
}

func getErrForField(e error, path string) string {
	var errMap map[string]interface{}
	err := json.Unmarshal([]byte(e.Error()), &errMap)
	if err != nil {
		panic(err)
	}
	elements := strings.Split(path, ".")
	for i, element := range elements {
		v, ok := errMap[element]
		if !ok {
			return ""
		}
		if i+1 == len(elements) {
			return v.(string)
		}

		errMap, ok = v.(map[string]interface{})
		if !ok {
			panic(ok)
		}
	}
	return ""
}

func BenchmarkValidator_Validate(b *testing.B) {
	v, err := NewValidator()
	assert.Nil(b, err)
	schema, err := ioutil.ReadFile("testdata/good_schema.lua")
	assert.Nil(b, err)
	assert.Nil(b, v.LoadSchema(string(schema)))
	b.ResetTimer()
	plugin := `{
  "name": "key-auth",
  "config": {
    "foo": "bar",
    "key_names": "broken-on-purpose",
    "key_in_body": true
  }
}`
	for i := 0; i < b.N; i++ {
		err := v.Validate(plugin)
		assert.NotNil(b, err)
	}
}
