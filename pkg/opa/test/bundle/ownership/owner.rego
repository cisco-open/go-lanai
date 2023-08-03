# +short        current user is the owner of resource
# +desc         Checks for resource's "owner_id" against  current user
# +unknowns     input.resource.owner_id   # list of unknown inputs that this policy support
package ownership

# Check Ownership
is_owner {
    input.resource.owner_id == input.auth.user_id
}
