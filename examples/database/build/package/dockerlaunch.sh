#!/bin/bash

if [ -z "$POPULATE" ]; then
    exec ./skeleton-service "$@"
elif [ "$POPULATE" = "database" ]; then
    exec ./migrate "$@"
fi
