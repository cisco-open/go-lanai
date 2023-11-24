# +short        client is authorized to access actuator endpoint
# +desc         Example of checking if client is allowed to access API based on scopes
package actuator

import data.oauth2.has_scope

# +desc alive is public
allow_endpoint {
    input.request.endpoint_id = "alive"
}

# +desc health is public
allow_endpoint {
    input.request.endpoint_id = "health"
}

# +desc info is public
allow_endpoint {
    input.request.endpoint_id = "info"
}

# +desc the rest endpoints are protected by scope
allow_endpoint {
    has_scope("admin")
}
