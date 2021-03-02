# Go Proxy

[![License: MIT](https://img.shields.io/github/license/livesport-tv/goproxy)](https://opensource.org/licenses/MIT)
[![tests](https://github.com/livesport-tv/goproxy/actions/workflows/tests.yml/badge.svg)](https://github.com/livesport-tv/goproxy/actions/workflows/tests.yml)
[![build](https://github.com/livesport-tv/goproxy/actions/workflows/build.yml/badge.svg)](https://github.com/livesport-tv/goproxy/actions/workflows/build.yml)
[![Latest release](https://img.shields.io/github/v/release/livesport-tv/goproxy?display_name=tag&sort=semver)](https://github.com/livesport-tv/goproxy/releases)

Open source implementation of Go proxy with monorepo support (more projects with `go.mod` and different versions in the same GIT repository).

## Additional endpoints
This is an extension of basic Go proxy functionality.

| Address                                                             | Description                         |
|---------------------------------------------------------------------|-------------------------------------|
| `/versions.json`                                                    | Latest versions of modules.         |
| `/dl/{name}/{version}/{arch}`                                       | Downloads endpoint.                 |
| `/dl/versions.json`                                                 | Downloads latest versions.          |

Note: Downloads prefix (`dl`) is configurable.

## Configuration

| JSON path               | Description                                           | Example                       |
|-------------------------|-------------------------------------------------------|-------------------------------|
| `/addr`                 | Service HTTP listen address.                          | `":80"`                       |
| `/storage`              | Path to storage.                                      | `"./cache"`                   |
| `/log_level`            | Log level.                                            | `"trace"`                     |
| `/default_go_proxy_url` | URL of default Go proxy for fallback.                 | `"http://proxy.golang.org"`   |
| `/downloads_prefix`     | Prefix for downloads path.                            | `"dl"`                        |
| `/modules`              | [Modules configurations.](#modules-configuration)     |                               |
| `/downloads`            | [Downloads configurations.](#downloads-configuration) |                               |
| `/sources`              | [Sources configurations.](#sources-configuration)     |                               |

Available log levels are `panic`, `fatal`, `error`, `warn`, `info`, `debug`, `trace` or an empty string for default log level.

See local [configuration file](./example-config.json) for more details.

### Modules configuration

| JSON path               | Description                                        | Example                |
|-------------------------|----------------------------------------------------|------------------------|
| `/name`                 | Name of module without version suffix.             | `"example.com/go/lib"` |
| `/source`               | Source name from list of sources or `null`.        | `"gitlab-local"`       |
| `/source_params`        | Source parameters object (depends on source type). |                        |

The fallback to `default_go_proxy_url` can be disabled with the parameter `/source` set to `null`.
It results in 404 for a given module instead of the fallback to `default_go_proxy_url`, which could result in unexpected states confusing Go SDK.

As every request to get a module from Go SDK is a cascade of requests - an attempt to download `"example.com/go/lib"` is in fact done by a cascade of requests:
- `"example.com/go/lib"`
- `"example.com/go"`
- `"example.com"`

If the fallback for `"example.com/go"` and `"example.com"` wasn't disabled, these requests would be redirected to `default_go_proxy_url` and have finished with an error.

### Downloads configuration

| JSON path               | Description                                        | Example              |
|-------------------------|----------------------------------------------------|----------------------|
| `/mode`                 | Mode (only `generic-packages` is allowed).         | `"generic-packages"` |
| `/source`               | Source name from list of sources.                  | `"gitlab-local"`     |
| `/source_params`        | Source parameters object (depends on source type). |                      |

All configured downloads are available on path: `/dl/<name>/<version>` or `/dl/<name>/<version>/<arch>`

Version is `<major>.<minor>.<patch>` or `latest`.

### Sources configuration

| JSON path               | Description                                        | Example                       |
|-------------------------|----------------------------------------------------|-------------------------------|
| `/name`                 | Name of the source (unique at source list).        | `"gitlab-local"`              |
| `/type`                 | Source type from supported source types.           | `"gitlab"`                    |

#### Source type `gitlab`
Source configuration (at `/sources`):

| JSON path               | Description                                        | Example                        |
|-------------------------|----------------------------------------------------|--------------------------------|
| `/url`                  | URL of Gitlab.                                     | `"https://gitlab.example.com"` |
| `/auth`                 | Private token to access Gitlab.                    | `"1111111111"`                 |
| `/allow_insecure_tls`   | Do not fail on invalid certificate.                | `true`                         |

Source parameters configuration (at `/modules`):

| JSON path               | Description                                     | Example  |
|-------------------------|-------------------------------------------------|----------|
| `/project_id`           | Gitlab project ID.                              | `42`     |
| `/dir`                  | Directory with project relative to git root.    | `"lib"`  |
| `/tag_prefix`           | Tag prefix (e.g. `lib-` for tag `lib-v1.0.0`).  | `"lib-"` |
| `/version_dir`          | Each version at separated directory, see below. | `false`  |

If `/version_dir` is `true`, major versions 2 and higher are expected at subdirectories, e.g.:

| Versions            | Directory     |
|---------------------|---------------|
| `0.x.x` and `1.x.x` | `/project/`   |
| `2.x.x`             | `/project/v2` |
| `3.x.x`             | `/project/v3` |

Source downloads parameters configuration (at `/downloads`):

| JSON path               | Description                                        | Example   |
|-------------------------|----------------------------------------------------|-----------|
| `/project_id`           | Gitlab project ID.                                 | `42`      |
| `/package_name`         | Name of package.                                   | `"lib"`   |
| `/disable_architecture` | Remove `<arch>` parameter from URL.                | `false`   |
| `/file_extension`       | File extension at package registry (optional).     | `".yaml"` |

## File storage
- Root of file storage is configurable by `/storage` property in the config.
- Each module has its own directory (without version suffix `v2`).
- Each version has its own files with a version prefix (e.g. `v1.0.0`).
  * File `.lock` contains the date and time of lock.
    This file is present only if a new version is being processed.
  * File `.tmp` contains temporary data during version processing.
  * File `.info` is an info file (see Go proxy specification).
  * File `.mod` is Go modules file.
  * File `.zip` is zip archive with a whole module at specified version.

## Dockerfile
You must build this image from the root of the repository.
```shell script
docker build -t goproxy .
```

To set version:
```shell script
docker build --build-arg GOPROXY_VERSION=2.0.0 -t goproxy .
```
