FROM golang:1.13 as builder

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

WORKDIR /plugin/
COPY . .
RUN go get -v -d
RUN go build -a -tags netgo -o release/linux/amd64/drone-gitlab-releases

FROM plugins/base:multiarch

LABEL maintainer="Drone.IO Community <drone-dev@googlegroups.com>" \
  org.label-schema.name="Drone Gitlab Release" \
  org.label-schema.vendor="Drone.IO Community" \
  org.label-schema.schema-version="1.0"

COPY --from=builder /plugin/release/linux/amd64/drone-gitlab-releases /bin/

ENTRYPOINT ["/bin/drone-gitlab-releases"]
