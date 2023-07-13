package ownership

import future.keywords

# Check Sharing
is_shared(op) {
     op = input.resource.sharing[input.auth.user_id][_]
}

# Check Sharing Update
# sharing didn't change
allow_change_sharing {
    not input.resource.delta.sharing
}

# sharing is same as before
allow_change_sharing {
    input.resource.delta.sharing = input.resource.sharing
}

# owner can change sharing
allow_change_sharing {
    input.resource.owner_id = input.auth.user_id
}

# other user with proper shared status can change sharing
allow_change_sharing {
    input.resource.sharing[input.auth.user_id][_] = "share"
}