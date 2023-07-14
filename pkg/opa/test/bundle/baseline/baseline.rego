package baseline

allow {
    _ = input
}

filter {
    _ = input
    input.fail == false
}