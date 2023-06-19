package poc

import future.keywords
import data.tenancy.allow_tenant_access
import data.roles.has_permission
import data.ops.is
import data.ownership.is_owner
import data.ownership.is_shared

filter_read if {
    input.resource.type == "poc"
    is("read")
    has_permission("VIEW")
    allow_tenant_access
}

filter_read if {
    input.resource.type == "poc"
    is("read")
    is_owner
    allow_tenant_access
}

filter_read if {
    input.resource.type == "poc"
    is("read")
    is_shared("read")
    allow_tenant_access
}

filter_write if {
    input.resource.type == "poc"
    is("write")
    has_permission("MANAGE")
    allow_tenant_access
}

filter_write if {
    input.resource.type == "poc"
    is("write")
    is_owner
    allow_tenant_access
}

filter_write if {
    input.resource.type == "poc"
    is("write")
    is_shared("write")
    allow_tenant_access
}

filter_write if {
    input.resource.type == "poc"
    is("write")
    has_permission("MANAGE")
    allow_tenant_access
}

filter_delete if {
    input.resource.type == "poc"
    is("delete")
    has_permission("MANAGE")
    allow_tenant_access
}

filter_delete if {
    input.resource.type == "poc"
    is("delete")
    is_owner
    allow_tenant_access
}

filter_delete if {
    input.resource.type == "poc"
    is("delete")
    is_shared("delete")
    allow_tenant_access
}

