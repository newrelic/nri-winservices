# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## 1.0.1 (2023-04-04)
### Fixed
- Creation of WIN_SERVICE entities

## 1.0.0 (2023-03-16)
### Changed
- Bump golang.org/x/sys from 0.4.0 to 0.5.0
- Bump github.com/stretchr/testify from 1.8.1 to 1.8.2
- Bump github.com/prometheus/common from v0.37.0 to v0.42.0

## 0.6.0 Beta (2022-08-25)
### Changed
- Bumped exporter version to commit 1c199e6c0eed881fb09dfcc84eee191262215e5e fixing access to some restricted services on some windows distributions.

## 0.5.0 Beta (2021-09-28)
### Changed
- Spawned exporter process priority class is now set to the same as the integration

## 0.4.0 Beta (2021-09-27)
### Changed
- Fixed consistent memory increase in windows exporter

## 0.3.0 Beta (2021-08-26)
### Changed
- Improve performance by getting the Windows Services from the API instead of WMI
- Modify the metrics to the new V4 format
- Entities are identified as `WIN_SERVICE`

## 0.2.0 Beta (2021-02-19)
### Changed
- Timeout now defaults to 60s
- Regex example more clear

## 0.1.0 Beta (2020-06-17)
### Changed
- Make regex non case-sensitive

## 0.0.3 Alpha Release (2020-06-12)
### Added
- Initial release
