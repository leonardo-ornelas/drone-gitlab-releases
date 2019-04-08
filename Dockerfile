FROM plugins/base:multiarch

LABEL maintainer="Drone.IO Community <drone-dev@googlegroups.com>" \
  org.label-schema.name="Drone Gitlab Release" \
  org.label-schema.vendor="Drone.IO Community" \
  org.label-schema.schema-version="1.0"

COPY release/linux/amd64/drone-gitlab-releases /bin/

ENTRYPOINT ["/bin/drone-gitlab-releases"]
