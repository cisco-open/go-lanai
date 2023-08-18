# +short        current user is allowed to update sharing
# +desc         Checks for resource's "delta" and determine if sharing list of the resource is changed and such action is allowed by current user
# +unknowns     input.resource.sharing input.resource.delta.sharing  # list of unknown inputs that this policy support
package ownership

# Check Sharing Update
# +desc sharing didn't change
allow_change_sharing {
    not input.resource.delta.sharing
}

# +desc sharing is same as before
allow_change_sharing {
    input.resource.delta.sharing = input.resource.sharing
}

# +desc user is owner
allow_change_sharing {
    input.resource.owner_id = input.auth.user_id
}

# +desc user with proper shared status
allow_change_sharing {
    input.resource.sharing[input.auth.user_id][_] = "share"
}