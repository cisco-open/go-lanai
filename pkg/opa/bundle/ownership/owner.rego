package ownership

import future.keywords
import data.roles.has_permission

# Check Ownership
is_owner if {
    input.resource.owner_id == input.auth.user_id
}

# owner has permission
allow_change_owner if {
    has_permission("MANAGE")
}

# owner is same as current user
allow_change_owner if {
    input.resource.delta.owner_id == input.auth.user_id
}

# owner is not changed
allow_change_owner if {
    not input.resource.delta.owner_id
}