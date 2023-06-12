package ownership

import future.keywords

# Check shared status
default op := ""
is_shared(op) if {
     op = input.resource.share[input.auth.user_id][_]
}