![gitlab.logo](logo.svg)


# drone-gitlab-release

[![Join the discussion at https://discourse.drone.io](https://img.shields.io/badge/discourse-forum-orange.svg)](https://discourse.drone.io)
[![Drone questions at https://stackoverflow.com](https://img.shields.io/badge/drone-stackoverflow-orange.svg)](https://stackoverflow.com/questions/tagged/drone.io)

[![](https://images.microbadger.com/badges/image/alexgamas/drone-rocketchat.svg)](https://microbadger.com/images/alexgamas/drone-rocketchat "Get your own image badge on microbadger.com")

Drone plugin for creating a GitLab release. 

## Build

Build the binary with the following commands:

```
go build
```

## Docker

Build the Docker image with the following commands:

```
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo -o release/linux/amd64/drone-gitlab-releases
docker build --rm -t plugins/gitlab-release .
```

## Usage

Execute from the working directory:

```
docker run --rm \
  -e PLUGIN_ASSET=example.zip \
  -e PLUGIN_NAME="Release Name" \
  -e PLUGIN_TOKEN=gitLabToken \
  -e DRONE_BUILD_EVENT=tag \
  -e DRONE_REPO_LINK=https://gitlab.com/octocat/hello-world \
  -e DRONE_REPO=octocat/hello-world \
  -e DRONE_TAG=0.0.1 \
  plugins/gitlab-release
```
