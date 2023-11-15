package resource.user_model_ref

import future.keywords
import data.rbac.has_permission
import data.resource.is_type
import data.resource.is_op
import data.ownership.is_owner

# Filters
filter_read = allow_read
filter_delete = allow_delete

# Read
allow_read if {
    is_type("user_model_ref")
    is_op("read")
    has_permission("MANAGE")
}

allow_read if {
    is_type("user_model_ref")
    is_op("read")
    is_owner
}

# Create
allow_create if {
    is_type("user_model_ref")
    is_op("create")
    has_permission("MANAGE")
}

allow_create if {
    is_type("user_model_ref")
    is_op("create")
    is_owner
}

# Delete
allow_delete if {
    is_type("user_model_ref")
    is_op("delete")
    has_permission("MANAGE")
}

allow_delete if {
    is_type("user_model_ref")
    is_op("delete")
    is_owner
}


