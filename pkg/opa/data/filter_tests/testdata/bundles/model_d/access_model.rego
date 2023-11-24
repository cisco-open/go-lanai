package resource.model

import future.keywords
import data.tenancy.allow_tenant_access
import data.tenancy.allow_change_tenant
import data.rbac.has_permission
import data.resource.is_type
import data.resource.is_op

# Filters
filter_read = allow_read
filter_write = allow_write
filter_delete = allow_delete

# Read
allow_read if {
    is_type("model")
    is_op("read")
    allow_tenant_access
}

# Write/Update
# Note: If we allow someone to assign many2many relationships, we also need to allow user to update "updated_at" field
allow_write if {
    is_type("model")
    is_op("write")
    allow_tenant_access
    allow_change_tenant
}

# Create
allow_create if {
    is_type("model")
    is_op("create")
    has_permission("MANAGE")
    allow_tenant_access
}

# Delete
allow_delete if {
    is_type("model")
    is_op("delete")
    has_permission("MANAGE")
    allow_tenant_access
}
