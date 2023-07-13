package ownership

import future.keywords
import data.roles.has_permission

# Check Ownership
is_owner {
    input.resource.owner_id == input.auth.user_id
}

# Check Ownership Updates
# owner has permission
allow_change_owner {
    has_permission("MANAGE")
}

# owner is same as current user
allow_change_owner {
    input.resource.delta.owner_id == input.resource.owner_id
}

# owner is not changed
allow_change_owner {
    not input.resource.delta.owner_id
}