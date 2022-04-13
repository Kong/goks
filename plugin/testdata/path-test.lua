local typedefs = require "kong.db.schema.typedefs"

local schema = {
  name = "path-test",
  fields = {
    { config = {
        type = "record",
        fields = {
          { path =  typedefs.path{ required = true } }
        }
      }
    }
  }
}

return schema