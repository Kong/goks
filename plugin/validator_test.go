package plugin

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type KongPlugin struct {
	CreatedAt *int                   `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	ID        string                 `json:"id,omitempty" yaml:"id,omitempty"`
	Name      string                 `json:"name,omitempty" yaml:"name,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty" yaml:"config,omitempty"`
	Enabled   bool                   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Protocols []string               `json:"protocols,omitempty" yaml:"protocols,omitempty"`
	Tags      []string               `json:"tags,omitempty" yaml:"tags,omitempty"`
}

type KongPluginConfig struct {
	Fields []map[string]interface{} `json:"fields,omitempty" yaml:"fields,omitempty"`
}

func TestValidator_LoadSchema(t *testing.T) {
	v, err := NewValidator()
	assert.Nil(t, err)

	t.Run("loads a good schema", func(t *testing.T) {
		schema, err := ioutil.ReadFile("testdata/key-auth.lua")
		assert.Nil(t, err)
		assert.Nil(t, v.LoadSchema(string(schema)))
	})
	t.Run("loads a schema with entity checks", func(t *testing.T) {
		schema, err := ioutil.ReadFile("testdata/rate-limiting.lua")
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
	t.Run("fails to load a schema with invalid imports", func(t *testing.T) {
		schema, err := ioutil.ReadFile("testdata/invalid_import_schema.lua")
		assert.Nil(t, err)
		err = v.LoadSchema(string(schema))
		assert.NotNil(t, err)
	})
}

func TestValidator_Validate(t *testing.T) {
	v, err := NewValidator()
	assert.Nil(t, err)
	schema, err := ioutil.ReadFile("testdata/uuid_schema.lua")
	assert.Nil(t, err)
	assert.Nil(t, v.LoadSchema(string(schema)))
	schema, err = ioutil.ReadFile("testdata/key-auth.lua")
	assert.Nil(t, err)
	assert.Nil(t, v.LoadSchema(string(schema)))
	schema, err = ioutil.ReadFile("testdata/rate-limiting.lua")
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
	t.Run("errors out on unexpected keys", func(t *testing.T) {
		plugin := `{
                     "name": "key-auth",
                     "config": {
						"foo": "bar"
                     },
                     "enabled": true,
                     "protocols": [ "http"]
                   }`
		err := v.Validate(plugin)
		assert.Equal(t, `{"config":{"foo":"unknown field"}}`, err.Error())
	})
	t.Run("errors out when entity checkers fail", func(t *testing.T) {
		plugin := `{
                     "name": "rate-limiting",
                     "config": {},
                     "enabled": true,
                     "protocols": [ "http"]
                   }`
		err := v.Validate(plugin)
		assert.Equal(t, `{"@entity":["at least one of these fields `+
			`must be non-empty: 'config.second', 'config.minute', `+
			`'config.hour', 'config.day', 'config.month', 'config.year'"]}`,
			err.Error())
	})
	t.Run("errors out when multiple entity checkers fail", func(t *testing.T) {
		plugin := `{
                     "name": "rate-limiting",
                     "config": { "policy": "redis"},
                     "enabled": true,
                     "protocols": [ "http"]
                   }`
		err := v.Validate(plugin)
		expected := `{
  "@entity": [
    "at least one of these fields must be non-empty: 'config.second', ` +
			`'config.minute', 'config.hour', 'config.day', 'config.month', 'config.year'",
    "failed conditional validation given value of field 'config.policy'",
    "failed conditional validation given value of field 'config.policy'",
    "failed conditional validation given value of field 'config.policy'"
  ],
  "config": {
    "redis_host": "required field missing",
    "redis_port": "required field missing",
    "redis_timeout": "required field missing"
  }
}
`
		assert.JSONEq(t, expected, err.Error())
	})
}

func TestValidator_ProcessAutoFields(t *testing.T) {
	v, err := NewValidator()
	assert.Nil(t, err)
	schema, err := ioutil.ReadFile("testdata/key-auth.lua")
	assert.Nil(t, err)
	assert.Nil(t, v.LoadSchema(string(schema)))
	t.Run("populates defaults for key-auth plugin", func(t *testing.T) {
		plugin := `{
                     "name": "key-auth",
                     "config": {}
                   }`
		plugin, err := v.ProcessAutoFields(plugin)
		assert.Nil(t, err)

		var kongPlugin KongPlugin
		assert.Nil(t, json.Unmarshal([]byte(plugin), &kongPlugin))
		assert.LessOrEqual(t, *kongPlugin.CreatedAt, int(time.Now().Unix()))
		assert.NotEmpty(t, kongPlugin.ID)
		assert.Len(t, kongPlugin.ID, 36)
		assert.Equal(t, "key-auth", kongPlugin.Name)
		assert.ElementsMatch(t, []string{"http", "https", "grpc", "grpcs"},
			kongPlugin.Protocols)
		assert.Equal(t, false, kongPlugin.Config["hide_credentials"])
		assert.Equal(t, false, kongPlugin.Config["key_in_body"])
		assert.Equal(t, true, kongPlugin.Config["key_in_header"])
		assert.Equal(t, true, kongPlugin.Config["key_in_query"])
		assert.Equal(t, true, kongPlugin.Enabled)
		assert.Equal(t, []interface{}{"apikey"}, kongPlugin.Config["key_names"])
		assert.Nil(t, kongPlugin.Config["anonymous"])
		assert.Nil(t, kongPlugin.Config["consumer"])
		assert.Nil(t, kongPlugin.Config["service"])
		assert.Nil(t, kongPlugin.Config["route"])
		assert.Nil(t, kongPlugin.Config["tags"])
	})
}

func TestValidator_SchemaAsJSON(t *testing.T) {
	schemaNames := []string{
		"key-auth",
		"rate-limiting",
	}
	v, err := NewValidator()
	assert.Nil(t, err)
	for _, schemaName := range schemaNames {
		schema, err := ioutil.ReadFile("testdata/" + schemaName + ".lua")
		assert.Nil(t, err)
		assert.Nil(t, v.LoadSchema(string(schema)))
	}

	t.Run("returns a valid JSON schema for loaded plugin schema", func(t *testing.T) {
		for _, schemaName := range schemaNames {
			schema, err := v.SchemaAsJSON(schemaName)
			assert.Nil(t, err)
			var kongPluginConfig KongPluginConfig
			assert.Nil(t, json.Unmarshal([]byte(schema), &kongPluginConfig))
		}
	})

	t.Run("returns error when specifying unknown plugin", func(t *testing.T) {
		schema, err := v.SchemaAsJSON("invalid-plugin")
		assert.Error(t, err, "no plugin named 'invalid-plugin'")
		assert.Empty(t, schema)
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
