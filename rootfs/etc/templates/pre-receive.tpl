#!/bin/bash
strip_remote_prefix() {
    stdbuf -i0 -o0 -e0 sed "s/^/"$'\e[1G'"/"
}

GIT_HOME={{.GitHome}} \
SSH_CONNECTION="$SSH_CONNECTION" \
SSH_ORIGINAL_COMMAND="$SSH_ORIGINAL_COMMAND" \
REPOSITORY="$RECEIVE_REPO" \
USERNAME="$RECEIVE_USER" \
FINGERPRINT="$RECEIVE_FINGERPRINT" \
POD_NAMESPACE="$POD_NAMESPACE" \
boot git-receive | strip_remote_prefix
