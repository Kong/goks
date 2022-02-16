local typedefs        = require "kong.db.schema.typedefs"

local _M = {}

_M.config_schema = {
  type = "record",

  fields = {
    { host = typedefs.host },
    { port = typedefs.port },
  },

  entity_checks = {
    {
      mutually_required = { "host", "port" },
    },
  },
}

return _M
