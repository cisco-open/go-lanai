package actuator

# info is public
allow_endpoint {
    input.request.endpoint_id = "info"
}

# alive is public
allow_endpoint {
    input.request.endpoint_id = "alive"
}

# the rest are protected
allow_endpoint {
    input.request.endpoint_id = "env"
    input.auth.client.scopes[_] = "admin"
}

allow_health_details {
    input.auth.client.scopes[_] = "admin"
}