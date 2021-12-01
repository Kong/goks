package vm

import (
	"bytes"

	lua "github.com/yuin/gopher-lua"
)

var setuplua = `
local Errors = require "kong.db.errors"
local Entity = require "kong.db.schema.entity"
local inspect = require"inspect"
local typedefs = require "kong.db.schema.typedefs"

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

local plugin_loader = require("kong.db.schema.plugin_loader")
local _, err = plugin_loader.load_subschema(Plugins,"key-auth", errors)
_G["Plugins"] = Plugins
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
