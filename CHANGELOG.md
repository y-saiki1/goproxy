# Changelog

## [Unreleased]

## [1.0.4] - 2022-03-17
### Changed
- Updated:
  - `golang.org/x/sys`  to `v0.0.0-20220317061510-51cd9980dadf` (indirect dependency)

### Fixed
- MacOS build by updating `golang.org/x/sys`.

## [1.0.3] - 2022-03-17
### Changed
- Updated to Go 1.18.
- Updated:
  - `github.com/stretchr/testify` to `1.7.1`

## [1.0.2] - 2022-02-09
### Fixed
- App version did not propagate to Docker image.

## [1.0.1] - 2022-02-08
### Removed
- Unused `retract` directive from `go.mod`.

### Fixed
- Stack overflow error on the new version not found error.

## [1.0.0] - 2022-02-01
### Added
- First release of Go Proxy.

[Unreleased]: https://github.com/livesport-tv/goproxy/compare/v1.0.4...master
[1.0.4]: https://github.com/livesport-tv/goproxy/compare/v1.0.3...v1.0.4
[1.0.3]: https://github.com/livesport-tv/goproxy/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/livesport-tv/goproxy/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/livesport-tv/goproxy/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/livesport-tv/goproxy/releases/tag/v1.0.0
