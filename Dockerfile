FROM golang:1.21 as build

WORKDIR /go/artifactory_exporter
ADD . /go/artifactory_exporter

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

ARG VERSION
ARG SOURCE_COMMIT
ARG SOURCE_BRANCH
ARG BUILD_DATE
ARG BUILD_USER

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -o /go/bin/artifactory_exporter -ldflags " \
    -X github.com/prometheus/common/version.Version=${VERSION} \
    -X github.com/prometheus/common/version.Revision=${SOURCE_COMMIT} \
    -X github.com/prometheus/common/version.Branch=${SOURCE_BRANCH} \
    -X github.com/prometheus/common/version.BuildDate=${BUILD_DATE} \
    -X github.com/prometheus/common/version.BuildUser=${BUILD_USER}"

# Use distroless as minimal base image
# Refer to https://github.com/GoogleContainerTools/distroless for more details.
FROM gcr.io/distroless/static:nonroot
COPY --from=build /go/bin/artifactory_exporter /

USER 65532:65532
EXPOSE 9531

ENTRYPOINT ["/artifactory_exporter"]
