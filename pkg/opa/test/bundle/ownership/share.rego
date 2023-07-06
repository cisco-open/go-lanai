package ownership

import future.keywords

# Check Sharing
is_shared(op) if {
     op = input.resource.share[input.auth.user_id][_]
}

# Check Sharing Update
# sharing didn't change
allow_change_sharing if {
    not input.resource.delta.share
}

# sharing is same as before
allow_change_sharing if {
    input.resource.delta.share = input.resource.share
}

# owner can change sharing
allow_change_sharing if {
    input.resource.owner_id = input.auth.user_id
}

# other user with proper shared status can change sharing
allow_change_sharing if {
    input.resource.share[input.auth.user_id][_] = "share"
}