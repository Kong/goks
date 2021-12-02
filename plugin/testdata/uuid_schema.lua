local typedefs = require "kong.db.schema.typedefs"


return {
  name = "test",
  fields = {
    { consumer = typedefs.no_consumer },
    { protocols = typedefs.protocols_http },
    { config = {
        type = "record",
        fields = {
          { uuid = typedefs.uuid },
        },
     },
    },
},}
