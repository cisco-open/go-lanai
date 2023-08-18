package poc

import future.keywords
import data.tenancy.allow_tenant_access
import data.tenancy.allow_change_tenant
import data.rbac.has_permission
import data.resource.is_type
import data.resource.is_op
import data.ownership.is_owner
import data.ownership.is_shared
import data.ownership.allow_change_owner

# Filters
filter_read = allow_read
filter_write = allow_write
filter_delete = allow_delete

# Read
allow_read if {
    is_type("poc")
    is_op("read")
    has_permission("VIEW")
    allow_tenant_access
}

allow_read if {
    is_type("poc")
    is_op("read")
    is_owner
    allow_tenant_access
}

allow_read if {
    is_type("poc")
    is_op("read")
    is_shared("read")
    allow_tenant_access
}

# Write/Update
allow_write if {
    is_type("poc")
    is_op("write")
    has_permission("MANAGE")
    allow_tenant_access
    allow_change_owner
    allow_change_tenant
}

allow_write if {
    is_type("poc")
    is_op("write")
    is_owner
    allow_tenant_access
    allow_change_owner
    allow_change_tenant
}

allow_write if {
    is_type("poc")
    is_op("write")
    is_shared("write")
    allow_tenant_access
    allow_change_owner
    allow_change_tenant
}

# Create
allow_create if {
    is_type("poc")
    is_op("create")
    has_permission("MANAGE")
    is_owner
    allow_tenant_access
}

# Delete
allow_delete if {
    is_type("poc")
    is_op("delete")
    has_permission("MANAGE")
    allow_tenant_access
}

allow_delete if {
    is_type("poc")
    is_op("delete")
    is_owner
    allow_tenant_access
}

allow_delete if {
    is_type("poc")
    is_op("delete")
    is_shared("delete")
    allow_tenant_access
}

