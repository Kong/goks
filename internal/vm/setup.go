package vm

import (
	"bytes"

	lua "github.com/yuin/gopher-lua"
)

var setuplua = `
local json = require "go.json"
local Errors = require "kong.db.errors"
local Entity = require "kong.db.schema.entity"
local inspect = require"inspect"
local typedefs = require "kong.db.schema.typedefs"
local MetaSchema = require "kong.db.schema.metaschema"

local certificates_definition = require "kong.db.schema.entities.certificates"
assert(Entity.new(certificates_definition))

local services_definition = require "kong.db.schema.entities.services"
assert(Entity.new(services_definition))

local routes_definition = require "kong.db.schema.entities.routes"
assert(Entity.new(routes_definition))

local consumers_definition = require "kong.db.schema.entities.consumers"
assert(Entity.new(consumers_definition))

local plugins_definition = require "kong.db.schema.entities.plugins"
local Plugins = assert(Entity.new(plugins_definition))

local errors = Errors.new("goks")

_G["Plugins"] = Plugins
_G["validate"] = function(plugin)
  if type(plugin) ~= "string" then
    return nil, "invalidate type, expected string"
  end
  local tbl, err = json.decode(plugin)
  if err then
    return nil, err
  end
  local inspect = require "inspect"
  local ok ,err = Plugins:validate(tbl)
  if not ok then
    return nil, json.encode(err)
  end
  return nil, nil
end

_G["process_auto_fields"] = function(plugin)
  if type(plugin) ~= "string" then
    return nil, "invalidate type, expected string"
  end
  local tbl, err = json.decode(plugin)
  if err then
    return nil, err
  end
  local inspect = require "inspect"
  local p = Plugins:process_auto_fields(tbl, "insert")
  return json.encode(p)
end

_G["load_plugin_schema"] = function(plugin_schema_string)
  local plugin_schema = loadstring(plugin_schema_string)()
  local ok, err_t = MetaSchema.MetaSubSchema:validate(plugin_schema)
  if not ok then
    return nil, tostring(errors:schema_violation(err_t))
  end

  local plugin_name = plugin_schema.name
  ok, err = Entity.new_subschema(Plugins, plugin_name, plugin_schema)
  if not ok then
    return nil, "error initializing schema for plugin: " .. err
  end
end
`

func setup(l *lua.LState) error {
	buf := bytes.NewBufferString(setuplua)
	fn, err := l.Load(buf, "setup.lua")
	if err != nil {
		return err
	}
	l.Push(fn)
	err = l.PCall(0, lua.MultRet, nil)
	return err
}
