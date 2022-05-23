FROM golang:1.17 as build

WORKDIR /go/artifactory_exporter
ADD . /go/artifactory_exporter

RUN go get -d -v ./...

ARG VERSION
ARG SOURCE_COMMIT
ARG SOURCE_BRANCH
ARG BUILD_DATE
ARG BUILD_USER

RUN GOOS=linux GOARCH=amd64 go build -o /go/bin/artifactory_exporter -ldflags " \
    -X github.com/prometheus/common/version.Version=${VERSION} \
    -X github.com/prometheus/common/version.Revision=${SOURCE_COMMIT} \
    -X github.com/prometheus/common/version.Branch=${SOURCE_BRANCH} \
    -X github.com/prometheus/common/version.BuildDate=${BUILD_DATE} \
    -X github.com/prometheus/common/version.BuildUser=${BUILD_USER}"

FROM gcr.io/distroless/base-debian11
COPY --from=build /go/bin/artifactory_exporter /

USER   nobody
EXPOSE 9531

ENTRYPOINT ["./artifactory_exporter"]
