FROM golang:1.13 as build

WORKDIR /go/artifactory_exporter
ADD . /go/artifactory_exporter

RUN go get -d -v ./...

RUN GOOS=linux GOARCH=amd64 go build -o /go/bin/artifactory_exporter

FROM gcr.io/distroless/base-debian10
COPY --from=build /go/bin/artifactory_exporter /

USER   nobody
EXPOSE 9531

ENTRYPOINT ["./artifactory_exporter"]
