package plugin

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	pluginTesting "github.com/kong/goks/plugin/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

type KongPluginSchema struct {
	Fields []map[string]interface{} `json:"fields,omitempty" yaml:"fields,omitempty"`
}

func TestValidator_LoadSchema(t *testing.T) {
	v, err := NewValidator(nil)
	assert.Nil(t, err)

	t.Run("loads a good schema", func(t *testing.T) {
		schema, err := ioutil.ReadFile("testdata/key-auth.lua")
		assert.Nil(t, err)
		pluginName, err := v.LoadSchema(string(schema))
		assert.EqualValues(t, "key-auth", pluginName)
		assert.Nil(t, err)
	})
	t.Run("loads a schema with entity checks", func(t *testing.T) {
		schema, err := ioutil.ReadFile("testdata/rate-limiting.lua")
		assert.Nil(t, err)
		pluginName, err := v.LoadSchema(string(schema))
		assert.EqualValues(t, "rate-limiting", pluginName)
		assert.Nil(t, err)
	})
	t.Run("fails to load a bad schema", func(t *testing.T) {
		schema, err := ioutil.ReadFile("testdata/bad_schema.lua")
		assert.Nil(t, err)
		pluginName, err := v.LoadSchema(string(schema))
		assert.Empty(t, pluginName)
		assert.NotNil(t, err)
		expected := "name: field required for entity check"
		assert.True(t, strings.Contains(err.Error(), expected))
	})
	t.Run("fails to load a schema with invalid imports", func(t *testing.T) {
		schema, err := ioutil.ReadFile("testdata/invalid_import_schema.lua")
		assert.Nil(t, err)
		pluginName, err := v.LoadSchema(string(schema))
		assert.Empty(t, pluginName)
		assert.NotNil(t, err)
	})

	vInjected, err := NewValidator(&pluginTesting.LuaTree)
	assert.Nil(t, err)
	schema, err := ioutil.ReadFile("testdata/inject_filesystem_schema.lua")
	assert.Nil(t, err)
	t.Run("fails to load schema without injected file system - invalid imports", func(t *testing.T) {
		pluginName, err := v.LoadSchema(string(schema))
		assert.Empty(t, pluginName)
		assert.NotNil(t, err)
	})
	t.Run("load schema with injected file system that previously had invalid imports", func(t *testing.T) {
		pluginName, err := vInjected.LoadSchema(string(schema))
		assert.EqualValues(t, "inject-test", pluginName)
		assert.Nil(t, err)
	})
}

func TestValidator_Validate(t *testing.T) {
	v, err := NewValidator(nil)
	assert.Nil(t, err)
	schema, err := ioutil.ReadFile("testdata/uuid_schema.lua")
	assert.Nil(t, err)
	pluginName, err := v.LoadSchema(string(schema))
	assert.EqualValues(t, "test", pluginName)
	assert.Nil(t, err)
	schema, err = ioutil.ReadFile("testdata/key-auth.lua")
	assert.Nil(t, err)
	pluginName, err = v.LoadSchema(string(schema))
	assert.EqualValues(t, "key-auth", pluginName)
	assert.Nil(t, err)
	schema, err = ioutil.ReadFile("testdata/rate-limiting.lua")
	assert.Nil(t, err)
	pluginName, err = v.LoadSchema(string(schema))
	assert.EqualValues(t, "rate-limiting", pluginName)
	assert.Nil(t, err)
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

	v, err = NewValidator(&pluginTesting.LuaTree)
	assert.Nil(t, err)
	schema, err = ioutil.ReadFile("testdata/inject_filesystem_schema.lua")
	assert.Nil(t, err)
	pluginName, err = v.LoadSchema(string(schema))
	assert.EqualValues(t, "inject-test", pluginName)
	assert.Nil(t, err)
	t.Run("validates a plugin with injected import fields correctly", func(t *testing.T) {
		plugin := `{
                     "name": "inject-test",
                     "config": {
                       "shared": "bar"
                     },
                     "enabled": true,
                     "protocols": [ "http" ]
                   }`
		err := v.Validate(plugin)
		assert.NotNil(t, err)
		configErr := getErrForField(err, "config.shared")
		assert.Equal(t, "expected a record", configErr)

		plugin = `{
			"name": "inject-test",
			"config": {
				"shared": {
					"host": "bar"
				}
			},
			"enabled": true,
			"protocols": [ "http" ]
		}`
		err = v.Validate(plugin)
		assert.NotNil(t, err)
		expected := `{
			"config": {
				"shared": {
					"@entity": [
						"all or none of these fields must be set: 'host', 'port'"
					]
				}
			}
		}`
		require.JSONEq(t, expected, err.Error())

		plugin = `{
			"name": "inject-test",
			"config": {
				"shared": {
					"host": "bar",
					"port": 80
				}
			},
			"enabled": true,
			"protocols": [ "http" ]
		}`
		err = v.Validate(plugin)
		assert.Nil(t, err)
	})
}

func TestValidator_ProcessAutoFields(t *testing.T) {
	v, err := NewValidator(nil)
	assert.Nil(t, err)
	schema, err := ioutil.ReadFile("testdata/key-auth.lua")
	assert.Nil(t, err)
	pluginName, err := v.LoadSchema(string(schema))
	assert.EqualValues(t, "key-auth", pluginName)
	assert.Nil(t, err)
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
		"acl",
		"http-log",
		"key-auth",
		"rate-limiting",
		"udp-log",
	}
	v, err := NewValidator(nil)
	assert.Nil(t, err)
	for _, schemaName := range schemaNames {
		schema, err := ioutil.ReadFile("testdata/" + schemaName + ".lua")
		assert.Nil(t, err)
		pluginName, err := v.LoadSchema(string(schema))
		assert.EqualValues(t, schemaName, pluginName)
		assert.Nil(t, err)
	}

	t.Run("returns a valid JSON schema for loaded plugin schema", func(t *testing.T) {
		for _, schemaName := range schemaNames {
			schema, err := v.SchemaAsJSON(schemaName)
			assert.Nil(t, err)
			var kongPluginConfig KongPluginSchema
			assert.Nil(t, json.Unmarshal([]byte(schema), &kongPluginConfig))
		}
	})

	t.Run("validate schema for known good plugin [udp-log]", func(t *testing.T) {
		schema, err := v.SchemaAsJSON("udp-log")
		assert.Nil(t, err)
		var kongPluginConfig KongPluginSchema
		assert.Nil(t, json.Unmarshal([]byte(schema), &kongPluginConfig))
		assert.EqualValues(t, 2, len(kongPluginConfig.Fields))

		// Parse using DOM style with the map[string]interface{}
		// This is simpler than trying to stuff a dynamic array into a static
		// structure
		for _, field := range kongPluginConfig.Fields {
			if v := field["config"]; v != nil {
				config, ok := v.(map[string]interface{})
				assert.True(t, ok)
				assert.EqualValues(t, 3, len(config)) // required = true (auto field)
				assert.EqualValues(t, true, config["required"])
				assert.EqualValues(t, "record", config["type"])
				configFields, ok := config["fields"].([]interface{})
				assert.True(t, ok)
				assert.EqualValues(t, 4, len(configFields))
				for _, configField := range configFields {
					option, ok := configField.(map[string]interface{})
					assert.True(t, ok)
					if w := option["custom_fields_by_lua"]; w != nil {
						customFieldsByLua, ok := w.(map[string]interface{})
						assert.True(t, ok)
						assert.EqualValues(t, 3, len(customFieldsByLua))
						keys, ok := customFieldsByLua["keys"].(map[string]interface{})
						assert.True(t, ok)
						assert.EqualValues(t, 2, len(keys))
						values, ok := customFieldsByLua["values"].(map[string]interface{})
						assert.True(t, ok)
						assert.EqualValues(t, 2, len(keys))
						assert.EqualValues(t, 1, keys["len_min"])
						assert.EqualValues(t, "string", keys["type"])
						assert.EqualValues(t, "map", customFieldsByLua["type"])
						assert.EqualValues(t, 1, values["len_min"])
						assert.EqualValues(t, "string", values["type"])
					} else if w := option["host"]; w != nil {
						host, ok := w.(map[string]interface{})
						assert.True(t, ok)
						assert.EqualValues(t, 2, len(host))
						assert.EqualValues(t, true, host["required"])
						assert.EqualValues(t, "string", host["type"])
					} else if w := option["port"]; w != nil {
						port, ok := w.(map[string]interface{})
						assert.True(t, ok)
						assert.EqualValues(t, 3, len(port))
						assert.EqualValues(t, true, port["required"])
						assert.EqualValues(t, "integer", port["type"])
						assert.ElementsMatch(t, []float64{0, 65535}, port["between"])
					} else if w := option["timeout"]; w != nil {
						timeout, ok := w.(map[string]interface{})
						assert.True(t, ok)
						assert.EqualValues(t, 2, len(timeout))
						assert.EqualValues(t, 10000, timeout["default"])
						assert.EqualValues(t, "number", timeout["type"])
					} else {
						assert.Fail(t, "invalid config.fields for udp-log")
					}
				}
			} else if v := field["protocols"]; v != nil {
				protocols, ok := v.(map[string]interface{})
				assert.True(t, ok)
				assert.EqualValues(t, 4, len(protocols))
				assert.ElementsMatch(t, []string{"grpc", "grpcs", "http", "https"},
					protocols["default"])
				if w := protocols["elements"]; w != nil {
					elements, ok := w.(map[string]interface{})
					assert.True(t, ok)
					assert.EqualValues(t, 2, len(elements))
					assert.ElementsMatch(t, []string{"grpc", "grpcs", "http", "https", "tcp", "tls", "udp"},
						elements["one_of"])
					assert.EqualValues(t, "string", elements["type"])
				} else {
					assert.Fail(t, "missing protocol.elements for udp-log")
				}
				assert.EqualValues(t, true, protocols["required"])
				assert.EqualValues(t, "set", protocols["type"])
			} else {
				assert.Fail(t, "invalid item in fields for udp-log")
			}
		}
	})

	t.Run("ensure functions are removed before returning schema", func(t *testing.T) {
		schema, err := v.SchemaAsJSON("acl")
		assert.Nil(t, err)
		var kongPluginConfig KongPluginSchema
		assert.Nil(t, json.Unmarshal([]byte(schema), &kongPluginConfig))
		assert.EqualValues(t, 3, len(kongPluginConfig.Fields))

		shorthandFieldsValidated := false
		for _, field := range kongPluginConfig.Fields {
			if v := field["config"]; v != nil {
				config, ok := v.(map[string]interface{})
				assert.True(t, ok)
				assert.EqualValues(t, 4, len(config)) // required = true (auto field)
				if w := config["shorthand_fields"]; w != nil {
					shorthandFields, ok := w.([]interface{})
					assert.True(t, ok)
					assert.EqualValues(t, 2, len(shorthandFields))
					for _, shorthandField := range shorthandFields {
						options, ok := shorthandField.(map[string]interface{})
						assert.True(t, ok)
						assert.EqualValues(t, 1, len(options))

						// func field should be removed from blacklist and whitelist
						for _, option := range options {
							x, ok := option.(map[string]interface{})
							assert.True(t, ok)
							assert.EqualValues(t, 2, len(x))
							assert.Nil(t, x["func"])
						}
					}
					shorthandFieldsValidated = true
				}
			}
		}
		assert.True(t, shorthandFieldsValidated, "failed to parse config.shorthand_fields")
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
