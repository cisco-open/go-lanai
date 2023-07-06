package model

import future.keywords
import data.roles.has_permission
import data.ops.is
import data.ownership.is_owner
import data.ownership.is_shared
import data.ownership.allow_change_sharing
import data.ownership.allow_change_owner

# Filters
filter_read = allow_read
filter_write = allow_write
filter_delete = allow_delete

# Read
allow_read if {
    input.resource.type == "model"
    is("read")
    has_permission("VIEW")
}

allow_read if {
    input.resource.type == "model"
    is("read")
    is_owner
}

allow_read if {
    input.resource.type == "model"
    is("read")
    is_shared("read")
}

# Write/Update
allow_write if {
    input.resource.type == "model"
    is("write")
    has_permission("MANAGE")
    allow_change_owner
    allow_change_sharing
}

allow_write if {
    input.resource.type == "model"
    is("write")
    is_owner
    allow_change_owner
    allow_change_sharing
}

allow_write if {
    input.resource.type == "model"
    is("write")
    is_shared("write")
    allow_change_owner
    allow_change_sharing
}

# Create
allow_create if {
    input.resource.type == "model"
    is("create")
    has_permission("MANAGE")
    is_owner
}

# Delete
allow_delete if {
    input.resource.type == "model"
    is("delete")
    has_permission("MANAGE")
}

allow_delete if {
    input.resource.type == "model"
    is("delete")
    is_owner
}

allow_delete if {
    input.resource.type == "model"
    is("delete")
    is_shared("delete")
}

