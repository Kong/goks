local typedefs = require "kong.db.schema.typedefs"

return {
  name = "methods-and-headers",
  fields = {
    { protocols = typedefs.protocols_http },
    { config = {
        type = "record",
        fields = {
          { methods = typedefs.methods },
          { headers = typedefs.headers },
        },
      },
    },
  },
}