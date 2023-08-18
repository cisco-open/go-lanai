package baseline

allow {
    _ = input
}

filter {
    _ = input
    input.fail = false
}

default allow_custom := false

allow_custom {
    _ = input.auth
    _ = input.auth.user_id
    not input.fail
}

allow_custom {
    not input.fail
    input.allow_no_auth
}

