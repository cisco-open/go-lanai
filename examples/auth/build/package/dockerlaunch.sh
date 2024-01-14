#!/bin/bash

if [ -z "$POPULATE" ]; then
    exec ./example-auth-service "$@"
elif [ "$POPULATE" = "database" ]; then
    exec ./migrate "$@"
fi
