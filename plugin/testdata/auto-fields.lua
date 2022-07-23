local typedefs = require "kong.db.schema.typedefs"

return {
  name = "auto-fields",
  fields = {
    { protocols = typedefs.protocols_http },
    { config = {
        type = "record",
        fields = {
          { string = { type = "string", auto = true, required = true } },
          { uuid = typedefs.uuid({ auto = true, required = true }) },
          { created_at = { type = "number", auto = true, required = true } },
          { updated_at = { type = "integer", auto = true, required = true } },
        },
      },
    },
  },
}
