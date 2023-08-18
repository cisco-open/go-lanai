# +short        resource is shared with current user
# +desc         Checks if current user is on sharing list and the requested operation is allowed
# +unknowns     input.resource.sharing   # list of unknown inputs that this policy support
package ownership

# Check Sharing
is_shared(op) {
     op = input.resource.sharing[input.auth.user_id][_]
}
