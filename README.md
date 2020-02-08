# JFrog Artifactory Exporter 

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/peimanja/artifactory_exporter/Build)](https://github.com/peimanja/artifactory_exporter/actions) [![Docker Build](https://img.shields.io/docker/cloud/build/peimanja/artifactory_exporter)](https://hub.docker.com/r/peimanja/artifactory_exporter/builds) [![Go Report Card](https://goreportcard.com/badge/github.com/peimanja/artifactory_exporter)](https://goreportcard.com/report/github.com/peimanja/artifactory_exporter)

A [Prometheus](https://prometheus.io) exporter for [JFrog Artifactory](https://jfrog.com/artifactory) stats. 


## Note

This exporter is under development and more metrics will be added later. Tested on Artifactory Enterprise and OSS version `6.16.0`.

## Authentication

The Artifactory exporter requires **admin** user and it supports multiple means of authentication. The following methods are supported:
  * Basic Auth
  * Bearer Token

### Basic Auth

Basic auth may be used by setting `ARTI_USERNAME` and `ARTI_PASSWORD` environment variables.

### Bearer Token

Artifactory access tokens may be used via the Authorization header by setting `ARTI_ACCESS_TOKEN` environment variable.

## Usage

### Binary

Download the binary for your operation system from [release](https://github.com/peimanja/artifactory_exporter/releases) page and run it:
```bash
$ ./artifactory_exporter <flags>
```

### Docker

Set the credentials in `env_file_name` and you can deploy this exporter using the [peimanja/artifactory_exporter](https://registry.hub.docker.com/r/peimanja/artifactory_exporter/) Docker image:
:

```bash
$ docker run --env-file=env_file_name -p 9531:9531 peimanja/artifactory_exporter:latest <flags>
```

### Flags

```bash
$  docker run peimanja/artifactory_exporter:latest -h
usage: main --artifactory.user=ARTIFACTORY.USER [<flags>]

Flags:
  -h, --help                    Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":9531"
                                Address to listen on for web interface and telemetry.
      --web.telemetry-path="/metrics"
                                Path under which to expose metrics.
      --artifactory.scrape-uri="http://localhost:8081/artifactory"
                                URI on which to scrape JFrog Artifactory.
      --artifactory.ssl-verify  Flag that enables SSL certificate verification for the scrape URI
      --artifactory.timeout=5s  Timeout for trying to get stats from JFrog Artifactory.
      --log.level=info          Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt       Output format of log messages. One of: [logfmt, json]
```

| Flag / Environment Variable | Required | Default | Description |
| --------------------------- | -------- | ------- | ----------- |
| `web.listen-address`<br/>`WEB_LISTEN_ADDR` | No | `:9531`| Address to listen on for web interface and telemetry. |
| `web.telemetry-path`<br/>`WEB_TELEMETRY_PATH` | No | `/metrics` | Path under which to expose metrics. |
| `artifactory.scrape-uri`<br/>`ARTI_SCRAPE_URI` | No | `http://localhost:8081/artifactory` | URI on which to scrape JFrog Artifactory. |
| `artifactory.ssl-verify`<br/>`ARTI_SSL_VERIFY` | No | `true` | Flag that enables SSL certificate verification for the scrape URI. |
| `artifactory.timeout`<br/>`ARTI_TIMEOUT` | No | `5s` | Timeout for trying to get stats from JFrog Artifactory. |
| `log.level` | No | `info` | Only log messages with the given severity or above. One of: [debug, info, warn, error]. |
| `log.format` | No | `logfmt` | Output format of log messages. One of: [logfmt, json]. |
| `ARTI_USERNAME` | *No | | User to access Artifactory |
| `ARTI_PASSWORD` | *No | | Password of the user accessing the Artifactory |
| `ARTI_ACCESS_TOKEN` | *No | | Access token for accessing the Artifactory |

* Either `ARTI_USERNAME` and `ARTI_PASSWORD` or `ARTI_ACCESS_TOKEN` environment variables has to be set.

### Metrics

Some metrics are not available with Artifactory OSS license. The exporter returns the following metrics:

| Metric | Description | Labels | OSS support |
| ------ | ----------- | ------ | ------ |
| artifactory_up | Was the last scrape of Artifactory successful. |  | &#9989; |
| artifactory_exporter_total_scrapes | Current total artifactory scrapes. |  | &#9989; |
| artifactory_exporter_json_parse_failures |Number of errors while parsing Json. |  | &#9989; |
| artifactory_replication_enabled | Replication status for an Artifactory repository (1 = enabled). | `name`, `type`, `cron_exp` | |
| artifactory_security_groups | Number of Artifactory groups. | | |
| artifactory_security_users | Number of Artifactory users for each realm. | `realm` | |
| artifactory_storage_artifacts | Total artifacts count stored in Artifactory. |  | &#9989; |
| artifactory_storage_artifacts_size_bytes | Total artifacts Size stored in Artifactory in bytes. |  | &#9989; |
| artifactory_storage_binaries | Total binaries count stored in Artifactory. |  | &#9989; |
| artifactory_storage_binaries_size_bytes | Total binaries Size stored in Artifactory in bytes. |  | &#9989; |
| artifactory_storage_filestore_bytes | Total space in the file store in bytes. | `storage_dir`, `storage_type` | &#9989; |
| artifactory_storage_filestore_used_bytes | Space used in the file store in bytes. | `storage_dir`, `storage_type` | &#9989; |
| artifactory_storage_filestore_free_bytes | Space free in the file store in bytes. | `storage_dir`, `storage_type` | &#9989; |
| artifactory_storage_repo_used_bytes | Space used by an Artifactory repository in bytes. | `name`, `package_type`, `type` | &#9989; |
| artifactory_storage_repo_folders | Number of folders in an Artifactory repository. | `name`, `package_type`, `type` | &#9989; |
| artifactory_storage_repo_files | Number files in an Artifactory repository. | `name`, `package_type`, `type` | &#9989; |
| artifactory_storage_repo_items | Number Items in an Artifactory repository. | `name`, `package_type`, `type` | &#9989; |
| artifactory_storage_repo_percentage | Percentage of space used by an Artifactory repository. | `name`, `package_type`, `type` | &#9989; |
| artifactory_system_healthy | Is Artifactory working properly (1 = healthy). | | &#9989; |
| artifactory_system_license | License type and expiry as labels. | `type`, `licensed_to`, `expires` | &#9989; |
| artifactory_system_version | Version and revision of Artifactory as labels. | `version`, `revision` | &#9989; |
