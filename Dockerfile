FROM golang:1.21 as build

WORKDIR /go/artifactory_exporter
ADD . /go/artifactory_exporter

ARG VERSION
ARG SOURCE_COMMIT
ARG SOURCE_BRANCH
ARG BUILD_DATE
ARG BUILD_USER

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/artifactory_exporter -ldflags " \
    -X github.com/prometheus/common/version.Version=${VERSION} \
    -X github.com/prometheus/common/version.Revision=${SOURCE_COMMIT} \
    -X github.com/prometheus/common/version.Branch=${SOURCE_BRANCH} \
    -X github.com/prometheus/common/version.BuildDate=${BUILD_DATE} \
    -X github.com/prometheus/common/version.BuildUser=${BUILD_USER}"

FROM alpine:3.19
COPY --from=build /go/bin/artifactory_exporter /

USER   nobody
EXPOSE 9531

ENTRYPOINT ["/artifactory_exporter"]
