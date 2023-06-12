package ownership

import future.keywords

# Check Ownership
#default is_owner := false
is_owner if {
    input.resource.owner_id == input.auth.user_id
}