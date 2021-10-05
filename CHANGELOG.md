# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## 0.5.0 Beta (2021-09-28)
### Changed
- Spawned exporter process priority class is now set to the same as the integration

## 0.4.0 Beta (2021-09-27)
### Changed
- Consistent memory increase in windows exporter

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
