local typedefs = require "kong.db.schema.typedefs"


local type = type


return {
  name        = "certificates",
  primary_key = { "id" },
  dao         = "kong.db.dao.certificates",
  workspaceable = true,

  fields = {
    { id         = typedefs.uuid, },
    { created_at = typedefs.auto_timestamp_s },
    { cert       = typedefs.certificate { required = true,  referenceable = true }, },
    { key        = typedefs.key         { required = true,  referenceable = true, encrypted = true }, },
    { cert_alt   = typedefs.certificate { required = false, referenceable = true }, },
    { key_alt    = typedefs.key         { required = false, referenceable = true, encrypted = true }, },
    { tags       = typedefs.tags },
  },

  entity_checks = {
    { mutually_required = { "cert_alt", "key_alt" } },
  }
}
