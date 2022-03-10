#!/usr/bin/env bash

VERSION=v3
CONTAINER="ghcr.io/hookerz/action-slack-notify:$VERSION"

# Update tag to latest commit
git push origin ":refs/tags/$VERSION"
git tag -f $VERSION
git push origin main --tags

# Build Docker container and deploy to Github packages
docker build --platform linux/amd64 -t action-slack-notify .
docker tag action-slack-notify $CONTAINER
docker push $CONTAINER
