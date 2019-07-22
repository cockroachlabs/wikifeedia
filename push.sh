#!/bin/bash
set -e

# Upload the working directory to be built.
BRANCH=$(git symbolic-ref --short HEAD)-$USER;
SHA=$(git rev-parse --short HEAD)-$USER;
gcloud --project cockroach-dev-inf builds submit --substitutions=BRANCH_NAME=$BRANCH,SHORT_SHA=$SHA

# Patch the running configuration; the key below is
# the name of the container since you can have multiple
# containers in a deployed pod.
kubectl set image deployment/wikifeedia wikifeedia=gcr.io/cockroach-dev-inf/cockroachlabs/wikifeedia:$SHA

echo "Now monitoring pod status. ^C to quit"
kubectl get po --watch --selector 'app=wikifeedia'

