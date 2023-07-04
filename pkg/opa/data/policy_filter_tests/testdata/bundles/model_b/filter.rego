package model

import future.keywords
import data.roles.has_permission
import data.ops.is
import data.ownership.is_owner
import data.ownership.is_shared
import data.ownership.allow_change_owner

filter_read if {
    input.resource.type == "model"
    is("read")
    has_permission("VIEW")
}

filter_read if {
    input.resource.type == "model"
    is("read")
    is_owner
}

filter_read if {
    input.resource.type == "model"
    is("read")
    is_shared("read")
}

filter_write if {
    input.resource.type == "model"
    is("write")
    has_permission("MANAGE")
    allow_change_owner
}

filter_write if {
    input.resource.type == "model"
    is("write")
    is_owner
    allow_change_owner
}

filter_write if {
    input.resource.type == "model"
    is("write")
    is_shared("write")
    allow_change_owner
}

filter_delete if {
    input.resource.type == "model"
    is("delete")
    has_permission("MANAGE")
}

filter_delete if {
    input.resource.type == "model"
    is("delete")
    is_owner
}

filter_delete if {
    input.resource.type == "model"
    is("delete")
    is_shared("delete")
}

