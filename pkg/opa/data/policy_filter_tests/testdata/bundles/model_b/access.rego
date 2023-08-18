package model

import future.keywords
import data.rbac.has_permission
import data.resource.is_type
import data.resource.is_op
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
    is_type("model")
    is_op("read")
    has_permission("VIEW")
}

allow_read if {
    is_type("model")
    is_op("read")
    is_owner
}

allow_read if {
    is_type("model")
    is_op("read")
    is_shared("read")
}

# Write/Update
allow_write if {
    is_type("model")
    is_op("write")
    has_permission("MANAGE")
    allow_change_owner
    allow_change_sharing
}

allow_write if {
    is_type("model")
    is_op("write")
    is_owner
    allow_change_owner
    allow_change_sharing
}

allow_write if {
    is_type("model")
    is_op("write")
    is_shared("write")
    allow_change_owner
    allow_change_sharing
}

# Create
allow_create if {
    is_type("model")
    is_op("create")
    has_permission("MANAGE")
    is_owner
}

# Delete
allow_delete if {
    is_type("model")
    is_op("delete")
    has_permission("MANAGE")
}

allow_delete if {
    is_type("model")
    is_op("delete")
    is_owner
}

allow_delete if {
    is_type("model")
    is_op("delete")
    is_shared("delete")
}

