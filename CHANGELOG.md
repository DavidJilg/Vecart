# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.1.1] - 2026-05-09

### Added

### Fixed

- Fixed invalid file existence check that led to files being unintentionally overwritten

### Changed

### Removed

## [2.1.0] - 2026-04-09

### Added

- Added "GCode" mode that allows to generate GCode for various CNC-machines based on svg files
- Added support for custom shapes defined in SVG files

### Fixed

- Fixed invalid parsing of camel case CLI arguments
- Fixed --debug CLI argument being interpreted as a config path
- Fixed height parsing of rectangle shapes
- Fixed filetypes of G-Code files being generated from batch configs
- Various other smaller bugfixes

### Changed

- Refactored the source code by splitting it into several internal packages

### Removed


## [2.0.0] - 2025-13-12

### Added

- Added new configuration options
  - **updateFrequency** | Specifies the update frequency for the console log in seconds.
  - **backgroundColor** | If set to other than "NONE" a colored background in the form of a rectangle covering the entire artboard is added.
  - **shortConfig** | Determines if only a shortend version of configuration is included as a comment in the SVG file.
  - **overwriteExisting**
- Added a new inbuild shape: 'heart'

### Fixed

- Various bug fixes and performance improvements.

### Changed

- Reworked the CLI interface for better usability (**breaking change!**).
- Reworked logging system for better clarity and more detailed information.

### Removed
