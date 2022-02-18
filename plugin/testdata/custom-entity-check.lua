local typedefs = require "kong.db.schema.typedefs"
return {
  name = "custom-entity-check",
  fields = {
    {protocols = typedefs.protocols},
    {config = {
        type = "record",
        fields = {
          {strategy = {type = "string", default = "localhost", one_of = { "localhost", "other"}}},
          {other = {type = "record", fields = {{host = typedefs.host}}}}
  }}}},
  entity_checks = {{
    custom_entity_check = {
      field_sources = { "config" },
      fn = function(entity)
        local config = entity.config
        if config.strategy == "other" and (config.other == nil or config.other.host == nil) then
          return nil, "custom-entity-check failed message"
        end
        return true
      end
  }}}
}