local utils = require "kong.tools.utils"
local typedefs = require "kong.db.schema.typedefs"

return {
  name = "uuid-custom",
  fields = {
    { protocols = typedefs.protocols },
    { config = {
        type = "record",
        fields = {
          { uuid = { type = "string", default = "", len_min = 0, }, },
        },
        custom_validator = function(config)
          return utils.is_valid_uuid(config.uuid)
        end,
      },
    },
  },
}
