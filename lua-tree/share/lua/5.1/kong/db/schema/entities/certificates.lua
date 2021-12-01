local typedefs = require "kong.db.schema.typedefs"


local type = type


return {
  name        = "certificates",
  primary_key = { "id" },
  dao         = "kong.db.dao.certificates",
  workspaceable = true,

  fields = {
    { id = typedefs.uuid, },
    { created_at     = typedefs.auto_timestamp_s },
    { cert           = typedefs.certificate { required = true }, },
    { key            = typedefs.key         { required = true }, },
    { cert_alt       = typedefs.certificate { required = false }, },
    { key_alt        = typedefs.key         { required = false }, },
    { tags           = typedefs.tags },
  },

  entity_checks = {
    { mutually_required = { "cert_alt", "key_alt" } },
  }
}
