package vm

import (
	"bytes"

	lua "github.com/yuin/gopher-lua"
)

var setuplua = `
local json = require "go.json"
local Errors = require "kong.db.errors"
local Entity = require "kong.db.schema.entity"
local inspect = require "inspect"
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
  local plugin_to_insert = Plugins:process_auto_fields(tbl, "insert")

  local ok, err = Plugins:validate(plugin_to_insert)
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
  local p = Plugins:process_auto_fields(tbl, "insert")
  return json.encode(p)
end

local function load_plugin_schema(plugin_schema_string, validate_plugins)
  local plugin_schema = loadstring(plugin_schema_string)()
  if plugin_schema == nil then
    return nil, "invalid plugin schema: cannot be empty"
  end
  local ok, err_t = MetaSchema.MetaSubSchema:validate(plugin_schema)
  if not ok then
    return nil, tostring(errors:schema_violation(err_t))
  end
  local plugin_name = plugin_schema.name
  local ok, err = Entity.new_subschema(validate_plugins, plugin_name, plugin_schema)
  if not ok then
    return nil, "error loading schema for plugin: " .. err
  end
  return plugin_schema, nil
end

_G["validate_plugin_schema"] = function(plugin_schema_string)
  local validate_plugins = assert(Entity.new(plugins_definition))
  local plugin_schema, err = load_plugin_schema(plugin_schema_string, validate_plugins)
  if err ~= nil then
    return nil, err
  end
  local plugin_name = plugin_schema.name
  validate_plugins.subschemas[plugin_name] = nil
  return plugin_name, nil
end

_G["load_plugin_schema"] = function(plugin_schema_string)
  local plugin_schema, err = load_plugin_schema(plugin_schema_string, Plugins)
  if err ~= nil then 
    return nil, err
  end
  return plugin_schema.name, nil
end

_G["unload_plugin_schema"] = function(plugin_name)
  local base_error = "error unloading schema for plugin: "
  if plugin_name == nil or plugin_name == "" then
    return false, base_error .. "plugin name must not be empty"
  end

  if Plugins.subschemas and Plugins.subschemas[plugin_name] then
    Plugins.subschemas[plugin_name] = nil
    return true, nil
  end
  return false, base_error .. "'" .. plugin_name .. "' does not exist"
end

-- Remove functions from a schema definition so that
-- cjson can encode the schema.
-- Taken from kong/api/api_helpers.lua
local schema_to_jsonable
do
  local insert = table.insert
  local ipairs = ipairs
  local next = next
  local cjson = { array_mt = {} } --- TODO(hbagdi) XXX analyze the impact

  local fdata_to_jsonable


  local function fields_to_jsonable(fields)
    local out = {}
    for _, field in ipairs(fields) do
      local fname = next(field)
      local fdata = field[fname]
      insert(out, { [fname] = fdata_to_jsonable(fdata, "no") })
    end
    setmetatable(out, cjson.array_mt)
    return out
  end


  -- Convert field data from schemas into something that can be
  -- passed to a JSON encoder.
  -- @tparam table fdata A Lua table with field data
  -- @tparam string is_array A three-state enum: "yes", "no" or "maybe"
  -- @treturn table A JSON-convertible Lua table
  fdata_to_jsonable = function(fdata, is_array)
    local out = {}
    local iter = is_array == "yes" and ipairs or pairs

    for k, v in iter(fdata) do
      if is_array == "maybe" and type(k) ~= "number" then
        is_array = "no"
      end

      if k == "schema" then
        out[k] = schema_to_jsonable(v)

      elseif type(v) == "table" then
        if k == "fields" and fdata.type == "record" then
          out[k] = fields_to_jsonable(v)

        elseif k == "default" and fdata.type == "array" then
          out[k] = fdata_to_jsonable(v, "yes")

        else
          out[k] = fdata_to_jsonable(v, "maybe")
        end

      elseif type(v) == "number" then
        if v ~= v then
          out[k] = "nan"
        elseif v == math.huge then
          out[k] = "inf"
        elseif v == -math.huge then
          out[k] = "-inf"
        else
          out[k] = v
        end

      elseif type(v) ~= "function" then
        out[k] = v
      end
    end
    if is_array == "yes" or is_array == "maybe" then
      setmetatable(out, cjson.array_mt)
    end
    return out
  end


  schema_to_jsonable = function(schema)
    local fields = fields_to_jsonable(schema.fields)
    return { fields = fields }
  end
	local schema_to_jsonable = schema_to_jsonable
end

-- Taken from kong/api/routes/kong.lua
local strip_foreign_schemas = function(fields)
  for _, field in ipairs(fields) do
    local fname = next(field)
    local fdata = field[fname]
    if fdata["type"] == "foreign" then
      fdata.schema = nil
    end
  end
end

_G["schema_as_json"] = function(schema_name)
  local subschema = Plugins.subschemas and Plugins.subschemas[schema_name] or nil
  if not subschema then
    return nil, "no plugin named '" .. schema_name .. "'"
  end
  local config = subschema.fields and subschema.fields.config or nil
  if not config then
    return nil, "no plugin schema available for '" .. schema_name .. "'"
  end

  local copy = schema_to_jsonable(subschema)
  strip_foreign_schemas(copy.fields)
  return json.encode(copy), nil
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
