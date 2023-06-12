package poc

import future.keywords
import data.tenancy.allow_tenant_access
import data.roles.has_permission
import data.ops.is
import data.ownership.is_owner
import data.ownership.is_shared

#default filter_read := false
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