# +entrypoint   allow_change_owner    # name of the policy to be queried
# +short        if changing resource's owner is allowed by current user
# +desc         Checks for resource's "delta" and determine if owner of the resource is changed and such action is allowed by current user
# +unknowns     input.resource.delta.owner_id input.resource.owner_id   # list of unknown inputs that this policy support
package ownership

import data.rbac.has_permission

# Check Ownership Updates
# +desc owner has permission
allow_change_owner {
    has_permission("MANAGE")
}

# +desc owner is same as current user
allow_change_owner {
    input.resource.delta.owner_id == input.resource.owner_id
}

# +desc owner is not changed
allow_change_owner {
    not input.resource.delta.owner_id
}