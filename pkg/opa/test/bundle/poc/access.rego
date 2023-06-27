package poc

import future.keywords
import data.tenancy.allow_tenant_access
import data.tenancy.allow_change_tenant
import data.roles.has_permission
import data.ops.is
import data.ownership.is_owner
import data.ownership.is_shared
import data.ownership.allow_change_owner

# Read
allow_read if {
    input.resource.type == "poc"
    is("read")
    has_permission("VIEW")
    allow_tenant_access
}

allow_read if {
    input.resource.type == "poc"
    is("read")
    is_owner
    allow_tenant_access
}

allow_read if {
    input.resource.type == "poc"
    is("read")
    is_shared("read")
    allow_tenant_access
}

# Write/Update
allow_write if {
    input.resource.type == "poc"
    is("write")
    has_permission("MANAGE")
    allow_tenant_access
    allow_change_owner
    allow_change_tenant
}

allow_write if {
    input.resource.type == "poc"
    is("write")
    is_owner
    allow_tenant_access
    allow_change_owner
    allow_change_tenant
}

allow_write if {
    input.resource.type == "poc"
    is("write")
    is_shared("write")
    allow_tenant_access
    allow_change_owner
    allow_change_tenant
}

# Create
allow_create if {
    input.resource.type == "poc"
    is("create")
    has_permission("MANAGE")
    allow_tenant_access
}

# Delete
allow_delete if {
    input.resource.type == "poc"
    is("delete")
    has_permission("MANAGE")
    allow_tenant_access
}

allow_delete if {
    input.resource.type == "poc"
    is("delete")
    is_owner
    allow_tenant_access
}

allow_delete if {
    input.resource.type == "poc"
    is("delete")
    is_shared("delete")
    allow_tenant_access
}

