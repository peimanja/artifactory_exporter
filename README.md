# JFrog Artifactory Exporter 

A [Prometheus](https://prometheus.io) exporter for [JFrog Artifactory](https://jfrog.com/artifactory) metrics. 


## Usage

### Docker

To run the firehose exporter as a Docker container, run:

```bash
$ docker run -p 9531:9531 --env ARTI_PASSWORD=$ARTI_ADMIN_PASS peimanja/artifactory_exporter:latest --artifactory.user=admin <flags>
```

### Flags

```bash
$  Docker run peimanja/artifactory_exporter:latest -h
usage: artifactory_exporter --artifactory.user=ARTIFACTORY.USER --artifactory.password=ARTIFACTORY.PASSWORD [<flags>]

Flags:
  -h, --help            Show context-sensitive help (also try --help-long and
                        --help-man).
      --web.listen-address=":9531"
                        Address to listen on for web interface and telemetry.
      --web.telemetry-path="/metrics"
                        Path under which to expose metrics.
      --artifactory.user=ARTIFACTORY.USER
                        User to access Artifactory.
      --artifactory.password=ARTIFACTORY.PASSWORD
                        Password of the user accessing the Artifactory.
      --artifactory.scrape-uri="http://localhost:8081/artifactory"
                        URI on which to scrape Artifactory.
      --artifactory.scrape-interval=30
                        How often to scrape Artifactory in secoonds.
      --exporter.debug  Enable debug mode.
```

| Flag / Environment Variable | Required | Default | Description |
| --------------------------- | -------- | ------- | ----------- |
| `web.listen-address`<br />`WEB_LISTEN_ADDR` | No | `:9531`| Address to listen on for web interface and telemetry |
| `web.telemetry-path`<br />`WEB_TELEMETRY_PATH` | No | `/metrics` | Path under which to expose Prometheus metrics |
| `artifactory.user`<br />`ARTI_USER` | Yes | | User to access Artifactory |
| `artifactory.password`<br />`ARTI_PASSWORD` | Yes | | Password of the user accessing the Artifactory |
| `artifactory.scrape-uri`<br />`ARTI_SCRAPE_URI` | No | `http://localhost:8081/artifactory` | JFrog Artifactory URL |
| `artifactory.scrape-interval`<br />`ARTI_SCRAPE_INTERVAL` | No | | JFrog Artifactory URL |
| `exporter.debug`<br />`DEBUG` | No | `false` | Enable debug mode |

### Metrics

The exporter returns the following metrics:

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| arti_up | Current health status of the server 1 = UP |  |
| arti_security_users | Number of artifactory users | `realm` |
| arti_artifacts_total_count | Total artifacts count stored in Artifactory |  |
| arti_artifacts_total_size_bytes | Total artifacts Size stored in Artifactory in bytes |  |
| arti_binaries_total_count | Total binaries count stored in Artifactory |  |
| arti_binaries_total_size_bytes | Total binaries Size stored in Artifactory in bytes |  |
| arti_filestore_bytes | Total space in the file store in bytes | `storage_dir`, `storage_type` |
| arti_filestore_used_bytes | Space used in the file store in bytes | `storage_dir`, `storage_type` |
| arti_filestore_free_bytes | Space free in the file store in bytes | `storage_dir`, `storage_type` |
| arti_repo_used_bytes | Space used by an Artifactory repository in bytes | `name`, `package_type`, `type` |
| arti_repo_folder_count | Number of folders in an Artifactory repository | `name`, `package_type`, `type` |
| arti_repo_files_count | Number files in an Artifactory repository | `name`, `package_type`, `type` |
| arti_repo_items_count | Number Items in an Artifactory repository | `name`, `package_type`, `type` |
| arti_repo_percentage | Percentage of space used by an Artifactory repository | `name`, `package_type`, `type` |
