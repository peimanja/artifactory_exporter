# JFrog Artifactory Exporter 

A [Prometheus](https://prometheus.io) exporter for [JFrog Artifactory](https://jfrog.com/artifactory) stats. 


## Note
This exporter is under development and more metrics will be added.


## Usage

### Docker

To run the firehose exporter as a Docker container, run:

```bash
$ docker run -p 9531:9531 --env ARTI_USERNAME=$ARTI_USERNAME --env ARTI_PASSWORD=$ARTI_ADMIN_PASS peimanja/artifactory_exporter:latest <flags>
```

### Flags

```bash
$  Docker run peimanja/artifactory_exporter:latest -h
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
| `artifactory.timeout`<br/>`ARTI_TIMEOUT` | No | `false` | Timeout for trying to get stats from JFrog Artifactory. |
| `ARTI_USERNAME` | Yes | | User to access Artifactory |
| `ARTI_PASSWORD` | Yes | | Password of the user accessing the Artifactory |

### Metrics

The exporter returns the following metrics:

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| artifactory_up | Was the last scrape of Artifactory successful. |  |
| artifactory_exporter_total_scrapes | Current total artifactory scrapes. |  |
| exporter_json_parse_failures |Number of errors while parsing Json. |  |
| artifactory_security_users | Number of Artifactory users for each realm. | `realm` |
| artifactory_storage_artifacts | Total artifacts count stored in Artifactory. |  |
| artifactory_storage_artifacts_size_bytes | Total artifacts Size stored in Artifactory in bytes. |  |
| artifactory_storage_binaries | Total binaries count stored in Artifactory. |  |
| artifactory_storage_binaries_size_bytes | Total binaries Size stored in Artifactory in bytes. |  |
| artifactory_storage_filestore_bytes | Total space in the file store in bytes. | `storage_dir`, `storage_type` |
| artifactory_storage_filestore_used_bytes | Space used in the file store in bytes. | `storage_dir`, `storage_type` |
| artifactory_storage_filestore_free_bytes | Space free in the file store in bytes. | `storage_dir`, `storage_type` |
| artifactory_storage_repo_used_bytes | Space used by an Artifactory repository in bytes. | `name`, `package_type`, `type` |
| artifactory_storage_repo_folders | Number of folders in an Artifactory repository. | `name`, `package_type`, `type` |
| artifactory_storage_repo_files | Number files in an Artifactory repository. | `name`, `package_type`, `type` |
| artifactory_storage_repo_items | Number Items in an Artifactory repository. | `name`, `package_type`, `type` |
| artifactory_storage_repo_percentage | Percentage of space used by an Artifactory repository. | `name`, `package_type`, `type` |
