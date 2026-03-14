# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2026-03-13

### Added
- Theme system with light/dark theme support
- Export functionality - Export hosts to JSON/YAML
- YAML configuration support - Full YAML config file support
- Host validation - Ping hosts and display online/offline status
- Quick connect with connection progress and error handling
- Profile system for connection settings

### Changed
- Improved host list with status indicators
- Enhanced error handling for SSH connections

### Fixed
- Various bug fixes and improvements

## [1.0.0] - 2024-03-04

### Added
- SSH config import - Parse and import hosts from `~/.ssh/config`
- Help/usage view - Press `?` to view all keyboard shortcuts
- Connection history tracking - Track connection attempts and view statistics
- Groups and tags for host organization
- Host details view with connection statistics

### Changed
- Improved UI styling with better color scheme
- Updated keyboard shortcuts in status bar
- Enhanced host list view with tag rendering

### Fixed
- Various bug fixes and improvements

## [0.1.0] - 2024-02-26

### Added
- Initial release
- Basic host management (add, edit, delete)
- Host list with search/filter
- SSH connection launching
- JSON-based persistent storage
