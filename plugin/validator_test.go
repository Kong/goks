package plugin

import (
	"encoding/json"
	"fmt"
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

func jsonify(v any) string {
	out, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(out)
}

func TestValidator_ValidateSchema(t *testing.T) {
	v, err := NewValidator(ValidatorOpts{})
	assert.Nil(t, err)

	t.Run("validate a valid schema", func(t *testing.T) {
		schema, err := ioutil.ReadFile("testdata/key-auth.lua")
		assert.Nil(t, err)
		pluginName, err := v.ValidateSchema(string(schema))
		assert.Nil(t, err)
		assert.EqualValues(t, "key-auth", pluginName)
	})

	t.Run("revalidate valid schemas multiple times", func(t *testing.T) {
		validPluginSchemas := []string{
			"acl",
			"all-typedefs",
			"between",
			"custom-entity-check",
			"http-log",
			"key-auth",
			"rate-limiting",
			"udp-log",
		}

		for _, validPluginSchema := range validPluginSchemas {
			for i := 0; i < 2; i++ {
				schema, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s.lua", validPluginSchema))
				assert.Nil(t, err)
				pluginName, err := v.ValidateSchema(string(schema))
				assert.Nil(t, err)
				assert.EqualValues(t, validPluginSchema, pluginName)
			}
		}
	})

	t.Run("fail validation against different schema problems", func(t *testing.T) {
		t.Run("bad schema", func(t *testing.T) {
			schema, err := ioutil.ReadFile("testdata/bad_schema.lua")
			assert.Nil(t, err)
			pluginName, err := v.ValidateSchema(string(schema))
			assert.NotNil(t, err)
			assert.Empty(t, pluginName)
			expected := "name: field required for entity check"
			assert.True(t, strings.Contains(err.Error(), expected))
		})

		t.Run("empty schema", func(t *testing.T) {
			pluginName, err := v.ValidateSchema("")
			assert.NotNil(t, err)
			assert.Empty(t, pluginName)
			expected := "invalid plugin schema: cannot be empty"
			assert.True(t, strings.Contains(err.Error(), expected))
		})

		t.Run("empty schema length > 0", func(t *testing.T) {
			pluginName, err := v.ValidateSchema("    ")
			assert.NotNil(t, err)
			assert.Empty(t, pluginName)
			expected := "invalid plugin schema: cannot be empty"
			assert.True(t, strings.Contains(err.Error(), expected))
		})
	})
}

func TestValidator_LoadSchema(t *testing.T) {
	v, err := NewValidator(ValidatorOpts{})
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

	vInjected, err := NewValidator(ValidatorOpts{InjectFS: &pluginTesting.LuaTree})
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

func TestValidator_ValidatePathGood(t *testing.T) {
	v, err := NewValidator(ValidatorOpts{})
	assert.Nil(t, err)
	schema, err := ioutil.ReadFile("testdata/path-test.lua")
	assert.Nil(t, err)
	pluginName, err := v.LoadSchema(string(schema))
	assert.EqualValues(t, "path-test", pluginName)
	assert.Nil(t, err)

	plugin := `{
	  "name": "path-test",
	  "config": {
		"path": "%s"
	  },
	  "enabled": true,
	  "protocols": ["http"]
	}`

	goodPaths := []string{
		"/",
		"/path",
		"/path/200",
		"/path/200/",
		"/path/200/test",
		"/path/200/test/",
		"/path/200/test-_~:@!$&'()*+,:=/",
		"/hello/path$with$!&'()*+,;=stuff",
	}
	for _, path := range goodPaths {
		err := v.Validate(fmt.Sprintf(plugin, path))
		assert.Nil(t, err)
	}
}

func TestValidator_ValidatePathBad(t *testing.T) {
	v, err := NewValidator(ValidatorOpts{})
	assert.Nil(t, err)
	schema, err := ioutil.ReadFile("testdata/path-test.lua")
	assert.Nil(t, err)
	pluginName, err := v.LoadSchema(string(schema))
	assert.EqualValues(t, "path-test", pluginName)
	assert.Nil(t, err)

	plugin := `{
		"name": "path-test",
		"config": {
		  "path": "%s"
		},
		"enabled": true,
		"protocols": ["http"]
	}`

	badPaths := []struct {
		name        string
		path        string
		expectedErr string
	}{
		{
			name: "errors out on empty path",
			path: "",
			expectedErr: `{
				"config": {
					"path": "length must be at least 1"
				}
			}`,
		},
		{
			name: "errors out on path not starting with /",
			path: "path",
			expectedErr: `{
				"config": {
					"path": "should start with: /"
				}
			}`,
		},
		{
			name: "errors out on path not starting with /",
			path: "path/200",
			expectedErr: `{
				"config": {
					"path": "should start with: /"
				}
			}`,
		},
		{
			name: "errors out on path containing not allowed elements",
			path: "/path/200?",
			expectedErr: `{
				"config": {
					"path": "invalid path: '/path/200?' (characters outside of the reserved list of RFC 3986 found)"
				}
			}`,
		},
		{
			name: "errors out on path containing not allowed elements",
			path: "/path/200#",
			expectedErr: `{
				"config": {
					"path": "invalid path: '/path/200#' (characters outside of the reserved list of RFC 3986 found)"
				}
			}`,
		},
		{
			name: "errors out on path containing not allowed elements",
			path: "/foo//bar",
			expectedErr: `{
				"config": {
					"path": "must not have empty segments"
				}
			}`,
		},
		{
			name: "errors out on path containing not allowed elements",
			path: "/foo/bar//",
			expectedErr: `{
				"config": {
					"path": "must not have empty segments"
				}
			}`,
		},
		{
			name: "errors out on path containing not allowed elements",
			path: "//foo/bar",
			expectedErr: `{
				"config": {
					"path": "must not have empty segments"
				}
			}`,
		},
		{
			name: "errors out on path containing not allowed elements",
			path: "[[/users/|foo/profile]]",
			expectedErr: `{
				"config": {
					"path": "should start with: /"
				}
			}`,
		},
		{
			name: "errors out on path containing not allowed elements",
			path: "[[/users/(this|foo/profile)]]",
			expectedErr: `{
				"config": {
					"path": "should start with: /"
				}
			}`,
		},
		{
			name: "errors out on path containing not allowed elements",
			path: "/some%2words",
			expectedErr: `{
				"config": {
					"path": "invalid url-encoded value: '%2w'"
				}
			}`,
		},
		{
			name: "errors out on path containing not allowed elements",
			path: "/some%0Xwords",
			expectedErr: `{
				"config": {
					"path": "invalid url-encoded value: '%0X'"
				}
			}`,
		},
		{
			name: "errors out on path containing not allowed elements",
			path: "/some%2Gwords",
			expectedErr: `{
				"config": {
					"path": "invalid url-encoded value: '%2G'"
				}
			}`,
		},
		{
			name: "errors out on path containing not allowed elements",
			path: "/some%20words%",
			expectedErr: `{
				"config": {
					"path": "invalid url-encoded value: '%'"
				}
			}`,
		},
		{
			name: "errors out on path containing not allowed elements",
			path: "/some%20words%a",
			expectedErr: `{
				"config": {
					"path": "invalid url-encoded value: '%a'"
				}
			}`,
		},
		{
			name: "errors out on path containing not allowed elements",
			path: "/some%20words%ax",
			expectedErr: `{
				"config": {
					"path": "invalid url-encoded value: '%ax'"
				}
			}`,
		},
	}
	for _, path := range badPaths {
		t.Run(path.name, func(t *testing.T) {
			err := v.Validate(fmt.Sprintf(plugin, path.path))
			assert.NotNil(t, err)
			require.JSONEq(t, path.expectedErr, err.Error())
		})
	}
}

func TestValidator_Validate(t *testing.T) {
	v, err := NewValidator(ValidatorOpts{})
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
	schema, err = ioutil.ReadFile("testdata/acl.lua")
	assert.Nil(t, err)
	pluginName, err = v.LoadSchema(string(schema))
	assert.EqualValues(t, "acl", pluginName)
	assert.Nil(t, err)
	schema, err = ioutil.ReadFile("testdata/udp-log.lua")
	assert.Nil(t, err)
	pluginName, err = v.LoadSchema(string(schema))
	assert.EqualValues(t, "udp-log", pluginName)
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
    "failed conditional validation given value of field 'config.policy'"
  ],
  "config": {
    "redis_host": "required field missing"
  }
}
`
		assert.JSONEq(t, expected, err.Error())
	})
	t.Run("tag validation succeeds with '-' in name", func(t *testing.T) {
		plugin := `{
      "name": "key-auth",
      "enabled": true,
      "protocols": [ "http"],
      "tags": [
        "i-have-hyphen",
        "idonthaveahyphen",
        "another-one-with-a-hyphen"
      ]
    }`
		err := v.Validate(plugin)
		assert.Nil(t, err)
	})
	t.Run("tag validation fails with ',' in name", func(t *testing.T) {
		plugin := `{
      "name": "key-auth",
      "enabled": true,
      "protocols": [ "http"],
      "tags": [
        "comma,fail"
      ]
    }`
		expectedErr := `{
			"tags": [
				"invalid tag 'comma,fail': expected printable ascii (except ` +
			"`,` and `/`" + `) or valid utf-8 sequences"
			]
		}`
		err := v.Validate(plugin)
		assert.NotNil(t, err)
		assert.JSONEq(t, expectedErr, err.Error())
	})
	t.Run("tag validation fails with '/' in name", func(t *testing.T) {
		plugin := `{
      "name": "key-auth",
      "enabled": true,
      "protocols": [ "http"],
      "tags": [
        "forward/fail"
      ]
    }`
		expectedErr := `{
			"tags": [
				"invalid tag 'forward/fail': expected printable ascii (except ` +
			"`,` and `/`" + `) or valid utf-8 sequences"
			]
		}`
		err := v.Validate(plugin)
		assert.NotNil(t, err)
		assert.JSONEq(t, expectedErr, err.Error())
	})
	t.Run("tag validation fails with multiple errors", func(t *testing.T) {
		plugin := `{
      "name": "key-auth",
      "enabled": true,
      "protocols": [ "http"],
      "tags": [
        "comma,fail",
        "forward/fail"
      ]
    }`
		expectedErr := `{
			"tags": [
				"invalid tag 'comma,fail': expected printable ascii (except ` +
			"`,` and `/`" + `) or valid utf-8 sequences",
				"invalid tag 'forward/fail': expected printable ascii (except ` +
			"`,` and `/`" + `) or valid utf-8 sequences"
			]
		}`
		err := v.Validate(plugin)
		assert.NotNil(t, err)
		assert.JSONEq(t, expectedErr, err.Error())
	})
	t.Run("only_one_of entity check passes", func(t *testing.T) {
		plugin := `{
      "name": "acl",
      "enabled": true,
      "protocols": [ "http"],
			"config": {
				"allow": [
					"foo"
				]
			}
    }`
		err := v.Validate(plugin)
		assert.Nil(t, err)

		plugin = `{
      "name": "acl",
      "enabled": true,
      "protocols": [ "http" ],
			"config": {
				"deny": [
					"foo"
				]
			}
    }`
		err = v.Validate(plugin)
		assert.Nil(t, err)
	})
	t.Run("only_one_of entity check fails", func(t *testing.T) {
		plugin := `{
      "name": "acl",
      "enabled": true,
      "protocols": [ "http"],
			"config": {
				"allow": [
					"foo"
				],
				"deny": [
					"bar"
				]
			}
    }`
		err := v.Validate(plugin)
		assert.NotNil(t, err)
		expectedErr := `{
			"@entity": [
				"exactly one of these fields must be non-empty: 'config.allow', 'config.deny'"
			]
		}`
		assert.JSONEq(t, expectedErr, err.Error())
	})
	t.Run("ensure all stream subsystem protocols can be assigned", func(t *testing.T) {
		protocols := []string{
			"tcp",
			"tls",
			"udp",
		}
		pluginFormat := `{
      "name": "udp-log",
      "enabled": true,
      "protocols": [ "%s" ],
			"config": {
				"host": "example.com",
				"port": 443
			}
    }`

		for _, protocol := range protocols {
			err := v.Validate(fmt.Sprintf(pluginFormat, protocol))
			assert.Nil(t, err)
		}
	})
	t.Run("ensure all http subsystem protocols can be assigned", func(t *testing.T) {
		protocols := []string{
			"http",
			"https",
			"grpc",
			"grpcs",
		}
		pluginFormat := `{
      "name": "acl",
      "enabled": true,
      "protocols": [ "%s" ],
			"config": {
				"allow": [
					"foo"
				]
			}
    }`

		for _, protocol := range protocols {
			err := v.Validate(fmt.Sprintf(pluginFormat, protocol))
			assert.Nil(t, err)
		}
	})

	v, err = NewValidator(ValidatorOpts{InjectFS: &pluginTesting.LuaTree})
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

func TestValidator_ValidateWithCustomEntityCheck(t *testing.T) {
	v, err := NewValidator(ValidatorOpts{})
	assert.Nil(t, err)
	schema, err := ioutil.ReadFile("testdata/custom-entity-check.lua")
	assert.Nil(t, err)
	pluginName, err := v.LoadSchema(string(schema))
	assert.Nil(t, err)
	assert.EqualValues(t, "custom-entity-check", pluginName)

	t.Run("fails and returns error message", func(t *testing.T) {
		plugin := `{
			"name": "custom-entity-check",
			"config": {"strategy": "other"},
			"enabled": true,
			"protocols": ["http"]
		}`
		err := v.Validate(plugin)
		assert.NotNil(t, err)
		expected := `{
			"@entity": [
				"custom-entity-check failed message"
			]
		}`
		require.JSONEq(t, expected, err.Error())
	})
}

func TestValidator_ValidateBetweenEntityCheck(t *testing.T) {
	v, err := NewValidator(ValidatorOpts{})
	assert.Nil(t, err)
	schema, err := ioutil.ReadFile("testdata/between.lua")
	assert.Nil(t, err)
	pluginName, err := v.LoadSchema(string(schema))
	assert.Nil(t, err)
	assert.EqualValues(t, "between", pluginName)

	tests := []struct {
		name        string
		values      string
		wantErr     bool
		expectedErr string
	}{
		{
			name:    "valid single value",
			values:  "[100]",
			wantErr: false,
		},
		{
			name:    "valid multiple values",
			values:  "[100, 200]",
			wantErr: false,
		},
		{
			name:        "invalid value fails - lower bound",
			values:      "[99]",
			wantErr:     true,
			expectedErr: `{"config":{"values":["value should be between 100 and 599"]}}`,
		},
		{
			name:        "invalid value fails - upper bound",
			values:      "[600]",
			wantErr:     true,
			expectedErr: `{"config":{"values":["value should be between 100 and 599"]}}`,
		},
		{
			name:        "mix valid and invalid values fails - lower bound",
			values:      "[99, 100]",
			wantErr:     true,
			expectedErr: `{"config":{"values":["value should be between 100 and 599"]}}`,
		},
		{
			name:    "fails with two invalid values",
			values:  "[99, 600]",
			wantErr: true,
			expectedErr: `{"config":{"values":[
				"value should be between 100 and 599",
				"value should be between 100 and 599"
			]}}`,
		},
		{
			name:    "fails with three invalid values",
			values:  "[99, 600, 700]",
			wantErr: true,
			expectedErr: `{"config":{"values":[
				"value should be between 100 and 599",
				"value should be between 100 and 599",
				"value should be between 100 and 599"
			]}}`,
		},
		{
			name:    "three valid values",
			values:  "[100, 200, 300]",
			wantErr: false,
		},
		// The following tests handle the sparse array error message encoding
		{
			name:        "mix valid and invalid value fails = upper bound",
			values:      "[100, 600]",
			wantErr:     true,
			expectedErr: `{"config":{"values":["[2] = value should be between 100 and 599"]}}`,
		},
		{
			name:    "fails with mix of valid and invalid values - all from sparse array",
			values:  "[100, 200, 300, 600, 500, 700]",
			wantErr: true,
			expectedErr: `{"config":{"values":[
				"[4] = value should be between 100 and 599",
				"[6] = value should be between 100 and 599"
			]}}`,
		},
		{
			name:    "fails with mix of valid and invalid values - indexed and sparse",
			values:  "[99, 200, 601, 600, 500, 300]",
			wantErr: true,
			expectedErr: `{"config":{"values":[
				"value should be between 100 and 599",
				"[3] = value should be between 100 and 599",
				"[4] = value should be between 100 and 599"
			]}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plugin := `{
				"name": "between",
				"enabled": true,
				"protocols": ["http"],
				"config": {
					"values": %s
				}
			}`
			err := v.Validate(fmt.Sprintf(plugin, test.values))
			if test.wantErr {
				assert.NotNil(t, err)
				require.JSONEq(t, test.expectedErr, err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidator_UnloadSchema(t *testing.T) {
	v, err := NewValidator(ValidatorOpts{})
	assert.Nil(t, err)

	pluginACL := `{
		"name": "acl",
		"enabled": true,
		"protocols": [ "http" ],
		"config": { "allow": [ "foo" ] }
	}`

	t.Run("ensure plugin schemas can be unloaded", func(t *testing.T) {
		// Ensure plugin schema is loaded and a configuration can be validated
		schema, err := ioutil.ReadFile("testdata/acl.lua")
		assert.Nil(t, err)
		pluginName, err := v.LoadSchema(string(schema))
		assert.Nil(t, err)
		assert.Equal(t, "acl", pluginName)
		err = v.Validate(pluginACL)
		assert.Nil(t, err)

		// Unload the schema and ensure failure when validating plugin configuration
		expectedErr := `{
			"name": "plugin 'acl' not enabled; add it to the 'plugins' configuration property"
		}`
		err = v.UnloadSchema(pluginName)
		assert.Nil(t, err)
		err = v.Validate(pluginACL)
		assert.NotNil(t, err)
		require.JSONEq(t, expectedErr, err.Error())
	})

	t.Run("ensure unloading of empty plugin schema return error", func(t *testing.T) {
		err := v.UnloadSchema("")
		assert.NotNil(t, err)
		assert.EqualError(t, err, "error unloading schema for plugin: plugin name must not be empty")
	})

	t.Run("ensure unloading of non-existent plugin schema return error", func(t *testing.T) {
		err := v.UnloadSchema("invalid-plugin-name")
		assert.NotNil(t, err)
		assert.EqualError(t, err, "error unloading schema for plugin: 'invalid-plugin-name' does not exist")
	})
}

func TestValidator_ProcessAutoFields(t *testing.T) {
	v, err := NewValidator(ValidatorOpts{})
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
	v, err := NewValidator(ValidatorOpts{})
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
					assert.ElementsMatch(t, []string{
						"grpc", "grpcs", "http", "https", "tcp", "tls", "tls_passthrough", "udp",
					},
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

func TestValidator_ValidateAllTypedefs(t *testing.T) {
	v, err := NewValidator(ValidatorOpts{})
	assert.Nil(t, err)
	schema, err := ioutil.ReadFile("testdata/all-typedefs.lua")
	assert.Nil(t, err)
	pluginName, err := v.LoadSchema(string(schema))
	assert.Nil(t, err)
	assert.EqualValues(t, "all-typedefs", pluginName)

	tests := []struct {
		name        string
		config      string
		wantErr     bool
		expectedErr string
	}{
		{
			name: "valid http method",
			config: `{
				"http_method": "GET"
			}`,
		},
		{
			name: "invalid http method",
			config: `{
				"http_method": "get"
			}`,
			expectedErr: `{"config":{"http_method":"invalid value: get"}}`,
			wantErr:     true,
		},
		{
			name: "valid protocol",
			config: `{
				"protocol": "http"
			}`,
		},
		{
			name: "invalid protocol",
			config: `{
				"protocol": "httpp"
			}`,
			expectedErr: `{"config":{
				"protocol":"expected one of: grpc, grpcs, http, https, tcp, tls, tls_passthrough, udp"
				}}`,
			wantErr: true,
		},
		{
			name: "valid host as ipv4",
			config: `{
				"host": "192.0.2.1"
			}`,
		},
		{
			name: "invalid host as ipv4",
			config: `{
				"host": "192.0.2.1.1"
			}`,
			expectedErr: `{"config":{"host":"invalid value: 192.0.2.1.1"}}`,
			wantErr:     true,
		},
		{
			name: "invalid host as ipv4:port",
			config: `{
				"host": "192.0.2.1:80"
			}`,
			expectedErr: `{"config":{"host":"must not have a port"}}`,
			wantErr:     true,
		},
		{
			name: "valid host as ipv6",
			config: `{
				"host": "2001:db8::1"
			}`,
		},
		{
			name: "invalid host as ipv6",
			config: `{
				"host": "2001:db8::1::1"
			}`,
			expectedErr: `{"config":{"host":"invalid value: 2001:db8::1::1"}}`,
			wantErr:     true,
		},
		{
			name: "valid host_with_optional_port as domain",
			config: `{
				"host_with_optional_port": "foo.com"
			}`,
		},
		{
			name: "valid host_with_optional_port as domain:port",
			config: `{
				"host_with_optional_port": "foo.com:80"
			}`,
		},
		{
			name: "valid host_with_optional_port as ip",
			config: `{
				"host_with_optional_port": "192.0.2.1"
			}`,
		},
		{
			name: "valid host_with_optional_port as ip:port",
			config: `{
				"host_with_optional_port": "192.0.2.1:80"
			}`,
		},
		{
			name: "valid host_with_optional_port as ipv6",
			config: `{
				"host_with_optional_port": "2001:DB8::1"
			}`,
		},
		{
			name: "valid host_with_optional_port as ipv6:port",
			config: `{
				"host_with_optional_port": "[2001:DB8::1]:80"
			}`,
		},
		{
			name: "invalid ipv4 in host_with_optional_port",
			config: `{
				"host_with_optional_port": "192.0.2.1.1:80"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"host_with_optional_port":"invalid ipv4 address: 192.0.2.1.1:80"}}`,
		},
		{
			name: "invalid ipv6 in host_with_optional_port",
			config: `{
				"host_with_optional_port": "2001:DB8::1::1"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"host_with_optional_port":"invalid ipv6 address: 2001:DB8::1::1"}}`,
		},
		{
			name: "invalid port in host_with_optional_port",
			config: `{
				"host_with_optional_port": "[2001:DB8::1]:99999999"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"host_with_optional_port":"invalid port number"}}`,
		},
		{
			name: "invalid domain in host_with_optional_port",
			config: `{
				"host_with_optional_port": "foo!.com"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"host_with_optional_port":"invalid hostname: foo!.com"}}`,
		},
		{
			name: "valid wildcard_host",
			config: `{
				"wildcard_host": "*test.com"
			}`,
		},
		{
			name: "valid wildcard_host",
			config: `{
				"wildcard_host": "test.*"
			}`,
		},
		{
			name: "invalid wildcard_host with multiple *",
			config: `{
				"wildcard_host": "*test.*"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"wildcard_host":"only one wildcard must be specified"}}`,
		},
		{
			name: "invalid wildcard_host with * in the middle",
			config: `{
				"wildcard_host": "new*test"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"wildcard_host":"wildcard must be leftmost or rightmost character"}}`,
		},
		{
			name: "valid ipv4",
			config: `{
				"ip": "192.0.2.1"
			}`,
		},
		{
			name: "invalid ipv4",
			config: `{
				"ip": "192.0.2.1.1"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"ip":"not an ip address: 192.0.2.1.1"}}`,
		},
		{
			name: "valid ipv6",
			config: `{
				"ip": "2001:db8::1"
			}`,
		},
		{
			name: "valid ipv6",
			config: `{
				"ip": "::1"
			}`,
		},
		{
			name: "valid ipv6",
			config: `{
				"ip": "fe80::1"
			}`,
		},
		{
			name: "invalid ipv6",
			config: `{
				"ip": "2001:db8::1::1"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"ip":"not an ip address: 2001:db8::1::1"}}`,
		},
		{
			name: "valid ip_or_cidr",
			config: `{
				"ip_or_cidr": "192.0.2.0/24"
			}`,
		},
		{
			name: "invalid ip_or_cidr with too big mask",
			config: `{
				"ip_or_cidr": "192.0.2.0/244"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"ip_or_cidr":"invalid ip or cidr range: '192.0.2.0/244'"}}`,
		},
		{
			name: "valid cidr_v4",
			config: `{
				"cidr_v4": "192.0.2.0/24"
			}`,
		},
		{
			name: "invalid cidr_v4 with too big mask",
			config: `{
				"cidr_v4": "192.0.2.0/244"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"cidr_v4":"invalid ipv4 cidr range: '192.0.2.0/244'"}}`,
		},
		{
			name: "first valid port",
			config: `{
				"port": 0
			}`,
		},
		{
			name: "last valid port",
			config: `{
				"port": 65535
			}`,
		},
		{
			name: "valid port",
			config: `{
				"port": 80
			}`,
		},
		{
			name: "invalid port as string",
			config: `{
				"port": "80"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"port":"expected an integer"}}`,
		},
		{
			name: "invalid port as float",
			config: `{
				"port": 80.1
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"port":"expected an integer"}}`,
		},
		{
			name: "invalid too big port",
			config: `{
				"port": 9999999999
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"port":"value should be between 0 and 65535"}}`,
		},
		{
			name: "invalid negative port",
			config: `{
				"port": -1
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"port":"value should be between 0 and 65535"}}`,
		},
		{
			name: "valid url",
			config: `{
				"url": "https://www.konghq.com"
			}`,
		},
		{
			name: "invalid url with relative path",
			config: `{
				"url": "/about/teams"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"url":"missing host in url"}}`,
		},
		{
			name: "valid header_name",
			config: `{
				"header_name": "x-header"
			}`,
		},
		{
			name: "invalid character in header_name",
			config: `{
				"header_name": "!header"
			}`,
			wantErr: true,
			expectedErr: `{"config":{"header_name":"bad header name '!header', ` +
				`allowed characters are A-Z, a-z, 0-9, '_', and '-'"}}`,
		},
		{
			name: "first valid timeout",
			config: `{
				"timeout": 0
			}`,
		},
		{
			name: "last valid timeout",
			config: `{
				"timeout": 2147483646
			}`,
		},
		{
			name: "valid timeout",
			config: `{
				"timeout": 10
			}`,
		},
		{
			name: "invalid negative timeout",
			config: `{
				"timeout": -1
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"timeout":"value should be between 0 and 2147483646"}}`,
		},
		{
			name: "invalid too big timeout",
			config: `{
				"timeout": 2147483647
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"timeout":"value should be between 0 and 2147483646"}}`,
		},
		{
			name: "valid uuid",
			config: `{
				"uuid": "23e623b0-ca23-11ec-9d64-0242ac120002"
			}`,
		},
		{
			name: "invalid uuid",
			config: `{
				"uuid": "abc"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"uuid":"expected a valid UUID"}}`,
		},
		{
			name: "invalid empty uuid",
			config: `{
				"uuid": ""
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"uuid":"expected a valid UUID"}}`,
		},
		{
			name: "invalid not-string uuid",
			config: `{
				"uuid": -1
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"uuid":"expected a string"}}`,
		},
		{
			name: "valid auto_timestamp_s",
			config: `{
				"auto_timestamp_s": 1234
			}`,
		},
		{
			name: "invalid number auto_timestamp_s",
			config: `{
				"auto_timestamp_s": 1.234
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"auto_timestamp_s":"expected an integer"}}`,
		},
		{
			name: "invalid string auto_timestamp_s",
			config: `{
				"auto_timestamp_s": "1234"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"auto_timestamp_s":"expected an integer"}}`,
		},
		{
			name: "valid auto_timestamp_ms",
			config: `{
				"auto_timestamp_ms": 1.234
			}`,
		},
		{
			name: "invalid auto_timestamp_ms",
			config: `{
				"auto_timestamp_ms": "1.234"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"auto_timestamp_ms":"expected a number"}}`,
		},
		{
			name: "valid name",
			config: `{
				"name": "foo.-_~"
			}`,
		},
		{
			name: "invalid name",
			config: `{
				"name": "foo.-_~!"
			}`,
			wantErr: true,
			expectedErr: `{"config":{
				"name":"invalid value 'foo.-_~!': it must only contain alphanumeric and '., -, _, ~' characters"
				}}`,
		},
		{
			name: "valid utf8_name",
			config: `{
				"utf8_name": "foo.-_~"
			}`,
		},
		{
			name: "invalid name with unaccepted ascii",
			config: `{
				"utf8_name": "fõõ.-_~"
			}`,
			wantErr: true,
			expectedErr: `{"config":{
				"utf8_name":"invalid value 'fõõ.-_~': the only accepted ascii characters are alphanumerics or ., -, _, and ~"
				}}`,
		},
		{
			name: "invalid name",
			config: `{
				"utf8_name": "foo.-_~!"
			}`,
			wantErr: true,
			expectedErr: `{"config":{
				"utf8_name":"invalid value 'foo.-_~!': the only accepted ascii characters are alphanumerics or ., -, _, and ~"
				}}`,
		},
		{
			name: "valid sni",
			config: `{
				"sni": "www.konghq.com"
			}`,
		},
		{
			name: "invalid sni with port",
			config: `{
				"sni": "www.konghq.com:8080"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"sni":"must not have a port"}}`,
		},
		{
			name: "invalid sni as IPv4",
			config: `{
				"sni": "192.0.2.0"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"sni":"must not be an IP"}}`,
		},
		{
			name: "invalid sni as IPv6",
			config: `{
				"sni": "2001:db8::1"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"sni":"must not be an IP"}}`,
		},
		{
			name: "valid certificate",
			config: jsonify(map [string]string{
				"certificate": pluginTesting.Cert,
			}),
		},
		{
			name: "valid ecdsa-with-SHA256 certificate",
			config: jsonify(map [string]string{
				"certificate": pluginTesting.CertEcdsa,
			}),
		},
		{
			name: "valid webclient certificate",
			config: jsonify(map [string]string{
				"certificate": pluginTesting.CertClient,
			}),
		},
		{
			name: "valid webclient2 certificate",
			config: jsonify(map [string]string{
				"certificate": pluginTesting.CertClient2,
			}),
		},
		{
			name: "expired certificate",
			config: jsonify(map [string]string{
				"certificate": pluginTesting.CertAlt,
			}),
		},
		{
			name: "valid alt ecdsa certificate",
			config: jsonify(map [string]string{
				"certificate": pluginTesting.CertAltEcdsa,
			}),
		},
		{
			name: "expired alt-alt certificate",
			config: jsonify(map [string]string{
				"certificate": pluginTesting.CertAltAlt,
			}),
		},
		{
			name: "valid alt-alt ecdsa certificate",
			config: jsonify(map [string]string{
				"certificate": pluginTesting.CertAltAltEcdsa,
			}),
		},
			{
			name: "valid CA certificate",
			config: jsonify(map [string]string{
				"certificate": pluginTesting.CertCA,
			}),
		},
		{
			name: "valid key",
			config: jsonify(map [string]string{
				"key": pluginTesting.Key,
			}),
		},
		{
			name: "valid ecdsa-with-SHA256 key",
			config: jsonify(map [string]string{
				"key": pluginTesting.KeyEcdsa,
			}),
		},
		{
			name: "valid webclient key",
			config: jsonify(map [string]string{
				"key": pluginTesting.KeyClient,
			}),
		},
		{
			name: "valid webclient2 key",
			config: jsonify(map [string]string{
				"key": pluginTesting.KeyClient2,
			}),
		},
		{
			name: "valid alt key",
			config: jsonify(map [string]string{
				"key": pluginTesting.KeyAlt,
			}),
		},
		{
			name: "valid alt ecdsa key",
			config: jsonify(map [string]string{
				"key": pluginTesting.KeyAltEcdsa,
			}),
		},
		{
			name: "valid alt-alt key",
			config: jsonify(map [string]string{
				"key": pluginTesting.KeyAltAlt,
			}),
		},
		{
			name: "valid alt-alt ecdsa key",
			config: jsonify(map [string]string{
				"key": pluginTesting.KeyAltAltEcdsa,
			}),
		},
		{
			name: "valid CA key",
			config: jsonify(map [string]string{
				"key": pluginTesting.KeyCA,
			}),
		},
		{
			name: "valid tag",
			config: `{
				"tag": "tag1"
			}`,
		},
		{
			name: "invalid tag with invalid character",
			config: `{
				"tag": "tag1,"
			}`,
			wantErr: true,
			expectedErr: `{"config":{
				"tag":"invalid tag 'tag1,': expected printable ascii (except ` + "`,` and `/`" +
				`) or valid utf-8 sequences"}}`,
		},
		{
			name: "valid tags",
			config: `{
				"tags": [
					"tag1"
				]
			}`,
		},
		{
			name: "valid tags",
			config: `{
				"tags": [
					"tag1",
					"tag2!"
				]
			}`,
		},
		{
			name: "invalid tags with invalid character",
			config: `{
				"tags": [
					"tag1",
					"tag2,"
				]
			}`,
			wantErr: true,
			expectedErr: `{"config":{"tags":["[2] = invalid tag 'tag2,': ` +
				`expected printable ascii (except ` + "`,` and `/`)" + ` or valid utf-8 sequences"]}}`,
		},
		{
			name: "invalid tags with invalid character",
			config: `{
				"tags": [
					"tag1",
					"tag2/"
				]
			}`,
			wantErr: true,
			expectedErr: `{"config":{"tags":["[2] = invalid tag 'tag2/': ` +
				`expected printable ascii (except ` + "`,` and `/`)" + ` or valid utf-8 sequences"]}}`,
		},
		{
			name: "invalid tags with ASCII utf8",
			config: `{
				"tags": [
					"tagõ"
				]
			}`,
			wantErr: true,
			expectedErr: `{"config":{"tags":["invalid tag 'tagõ': ` +
				`expected printable ascii (except ` + "`,` and `/`)" + ` or valid utf-8 sequences"]}}`,
		},
		{
			name: "valid sources with ip only",
			config: `{
				"sources": [
					{
						"ip": "192.0.2.1"
					},
					{
						"ip": "2001:db8::1"
					}
				]
			}`,
		},
		{
			name: "valid sources with port only",
			config: `{
				"sources": [
					{
						"port": 80
					}
				]
			}`,
		},
		{
			name: "valid sources with both ip and port",
			config: `{
				"sources": [
					{
						"ip": "192.0.2.1",
						"port": 80
					}
				]
			}`,
		},
		{
			name: "valid sources with multiple entries",
			config: `{
				"sources": [
					{
						"ip": "192.0.2.1",
						"port": 80
					},
					{
						"ip": "2001:db8::1"
					},
					{
						"port": 8080
					}
				]
			}`,
		},
		{
			name: "invalid sources with bogus ip",
			config: `{
				"sources": [
					{
						"ip": "192.0.2.1.1",
						"port": 80
					}
				]
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"sources":[{"ip": "invalid ip or cidr range: '192.0.2.1.1'"}]}}`,
		},
		{
			name: "invalid sources with bogus port",
			config: `{
				"sources": [
					{
						"ip": "192.0.2.1",
						"port": -80
					}
				]
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"sources":[{"port": "value should be between 0 and 65535"}]}}`,
		},
		{
			name: "valid destinations with ip only",
			config: `{
				"destinations": [
					{
						"ip": "192.0.2.1"
					},
					{
						"ip": "2001:db8::1"
					}
				]
			}`,
		},
		{
			name: "valid destinations with port only",
			config: `{
				"destinations": [
					{
						"port": 80
					}
				]
			}`,
		},
		{
			name: "valid destinations with both ip and port",
			config: `{
				"destinations": [
					{
						"ip": "192.0.2.1",
						"port": 80
					},
					{
						"ip": "2001:db8::1",
						"port": 80
					}
				]
			}`,
		},
		{
			name: "valid destinations with multiple entries",
			config: `{
				"destinations": [
					{
						"ip": "192.0.2.1",
						"port": 80
					},
					{
						"ip": "192.0.2.2"
					},
					{
						"port": 8080
					}
				]
			}`,
		},
		{
			name: "invalid destinations with bogus ipv4",
			config: `{
				"destinations": [
					{
						"ip": "192.0.2.1.1",
						"port": 80
					}
				]
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"destinations":[{"ip": "invalid ip or cidr range: '192.0.2.1.1'"}]}}`,
		},
		{
			name: "invalid destinations with bogus ipv6",
			config: `{
				"destinations": [
					{
						"ip": "2001:db8p::1",
						"port": 80
					}
				]
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"destinations":[{"ip": "invalid ip or cidr range: '2001:db8p::1'"}]}}`,
		},
		{
			name: "invalid destinations with bogus port",
			config: `{
				"destinations": [
					{
						"ip": "192.0.2.1",
						"port": -80
					}
				]
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"destinations":[{"port": "value should be between 0 and 65535"}]}}`,
		},
		{
			name: "valid hosts",
			config: `{
				"hosts": [
					"192.0.2.1",
					"www.konghq.com"
				]
			}`,
		},
		{
			name: "valid hosts with wildcards",
			config: `{
				"hosts": [
					"192.0.2.1",
					"*.konghq.com",
					"konghq.*"
				]
			}`,
		},
		{
			name: "valid hosts with wildcards",
			config: `{
				"hosts": [
					"192.0.2.1",
					"*konghq.com",
					"*",
					"konghq*",
					"*.konghq.*",
					"konghq.*.com"
				]
			}`,
			wantErr: true,
			expectedErr: `{"config":{"hosts":[
				"[2] = invalid wildcard: must be placed at leftmost or rightmost label",
				"[3] = invalid wildcard: must be placed at leftmost or rightmost label",
				"[4] = invalid wildcard: must be placed at leftmost or rightmost label",
				"[5] = invalid wildcard: must have at most one wildcard",
				"[6] = invalid wildcard: must be placed at leftmost or rightmost label"
				]}}`,
		},
		{
			name: "valid semantic_version",
			config: `{
				"semantic_version": "1"
			}`,
		},
		{
			name: "valid semantic_version",
			config: `{
				"semantic_version": "1.2"
			}`,
		},
		{
			name: "valid semantic_version",
			config: `{
				"semantic_version": "1.2.3"
			}`,
		},
		{
			name: "valid semantic_version",
			config: `{
				"semantic_version": "100.200.300"
			}`,
		},
		{
			name: "valid semantic_version",
			config: `{
				"semantic_version": "1.2.3-rc.1"
			}`,
		},
		{
			name: "valid semantic_version",
			config: `{
				"semantic_version": "1.2.3-alpha.1"
			}`,
		},
		{
			name: "valid semantic_version",
			config: `{
				"semantic_version": "1.2.3-beta.1"
			}`,
		},
		{
			name: "valid semantic_version",
			config: `{
				"semantic_version": "1.2.3.4-enterprise-edition"
			}`,
		},
		{
			name: "invalid semantic_version",
			config: `{
				"semantic_version": "hello"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"semantic_version":"invalid version number: must be in format of X.Y.Z"}}`,
		},
		{
			name: "invalid semantic_version",
			config: `{
				"semantic_version": ".1"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"semantic_version":"invalid version number: must be in format of X.Y.Z"}}`,
		},
		{
			name: "invalid semantic_version",
			config: `{
				"semantic_version": "1..1"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"semantic_version":"must not have empty version segments"}}`,
		},
		{
			name: "valid lua code gets accepted",
			config: `{
				"lua_code": {"header": "return nil"}
			}`,
		},
		{
			name: "invalid lua code gets rejected",
			config: `{
				"lua_code": "test"
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"lua_code":"expected a map"}}`,
		},
		{
			name: "invalid lua code syntax gets rejected",
			config: `{
				"lua_code": {"header": "hello"}
			}`,
			wantErr: true,
			expectedErr: `{"config":{"lua_code": "Error parsing function: ` +
				`lua-tree/share/lua/5.1/kong/tools/kong-lua-sandbox.lua:146: ` +
				`<string> at EOF:   parse error\n"}}`,
		},
		{
			name: "valid lua code but invalid function call doesn't return an error",
			config: `{
				"lua_code": {"header": "os.execute('echo hello')"}
			}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plugin := `{
				"name": "all-typedefs",
				"enabled": true,
				"protocols": ["http"],
				"config": %s
			}`
			err := v.Validate(fmt.Sprintf(plugin, test.config))
			if test.wantErr {
				assert.NotNil(t, err)
				require.JSONEq(t, test.expectedErr, err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidator_ValidateMethodsAndHeaders(t *testing.T) {
	v, err := NewValidator(ValidatorOpts{})
	assert.Nil(t, err)
	schema, err := ioutil.ReadFile("testdata/methods-and-headers.lua")
	assert.Nil(t, err)
	pluginName, err := v.LoadSchema(string(schema))
	assert.Nil(t, err)
	assert.EqualValues(t, "methods-and-headers", pluginName)

	tests := []struct {
		name        string
		config      string
		wantErr     bool
		expectedErr string
	}{
		{
			name: "valid using single header in array of headers",
			config: `{
				"headers": {
					"my-header": [
					  "value1"
					]
				}
			}`,
		},
		{
			name: "valid using multiple headers in array of headers",
			config: `{
				"headers": {
					"my-header": [
					  "value1"
					],
					"my-other-header": [
					  "value2",
					  "value3"
					]
				}
			}`,
		},
		{
			name: "invalid headers with invalid header name",
			config: `{
				"headers": {
					"my-header!": ["value1"]
				}
			}`,
			wantErr: true,
			expectedErr: `{"config":{
				"headers":"bad header name 'my-header!', allowed characters are A-Z, a-z, 0-9, '_', and '-'"
				}}`,
		},
		{
			name: "invalid headers with invalid header name as second header",
			config: `{
				"headers": {
					"my-valid-header": ["value1"],
					"my-invalid-header!": ["value1"]
				}
			}`,
			wantErr: true,
			expectedErr: `{"config":{
				"headers":"bad header name 'my-invalid-header!', allowed characters are A-Z, a-z, 0-9, '_', and '-'"
				}}`,
		},
		{
			name: "valid methods",
			config: `{
				"methods": ["GET", "POST"]
			}`,
		},
		{
			name: "invalid methods",
			config: `{
				"methods": ["GET", "post"]
			}`,
			wantErr:     true,
			expectedErr: `{"config":{"methods":["[2] = invalid value: post"]}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plugin := `{
				"name": "methods-and-headers",
				"enabled": true,
				"protocols": ["http"],
				"config": %s
			}`
			err := v.Validate(fmt.Sprintf(plugin, test.config))
			if test.wantErr {
				assert.NotNil(t, err)
				require.JSONEq(t, test.expectedErr, err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
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
			val, _ := v.(string)
			return val
		}

		errMap, ok = v.(map[string]interface{})
		if !ok {
			panic(ok)
		}
	}
	return ""
}
