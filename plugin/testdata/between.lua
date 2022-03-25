local typedefs = require "kong.db.schema.typedefs"

return {
  name = "between",
  fields = {
    { consumer = typedefs.no_consumer },
    { protocols = typedefs.protocols_http },
    { config = {
        type = "record",
        fields = {{
          values = {
            required = true,
            type = "array",
            elements = {
              type = "integer",
              between = { 100, 599 },
            },
          }},
        },
      },
    },
  },
}
