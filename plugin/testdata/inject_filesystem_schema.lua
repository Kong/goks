local typedefs = require "kong.db.schema.typedefs"
local shared = require "shared"
local constants = require "kong.constants"

return {
  name = constants.INJECT_PLUGIN_NAME,
  fields = {
    { consumer = typedefs.no_consumer },
    { protocols = typedefs.protocols_http },
    { config = {
        type = "record",
        fields = {
          { shared = shared.config_schema},
        },
      },
    },
  },
}
