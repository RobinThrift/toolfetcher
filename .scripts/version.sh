#!/bin/sh

version() {
    local version=`git describe --tags --abbrev=0 2>/dev/null`
    if [ -z "$version" ]; then
        echo "UNKNOWN-VERSION"
        exit 0;
    fi

    local num_commits=`git rev-list $version.. | wc -l | tr -d '[:space:]'`

    if [ "$num_commits" -eq "0" ]; then
        echo $version
    else
        echo $version\#$num_commits
    fi
}

version

