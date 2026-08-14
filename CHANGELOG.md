<!-- markdownlint-disable MD024 -->
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Deprecated

### Removed

### Fixed

### Security

## [0.3.0] - 2026-08-14

### Added

- **Plan-Time Asset Validation**: Added plan-time validation for data assets via backend dry-run validate endpoint during `terraform plan` on `masthead_data_product` resources
- **Configurable Request Timeout**: Added `request_timeout_seconds` provider configuration parameter to configure HTTP request timeout (default 60s)
- **Thread-Safe In-Memory Cache**: Implemented thread-safe in-memory caching and bulk warm-up for data domains and data products in the provider client to minimize redundant API roundtrips

### Changed

- **Set Semantics for Data Assets**: Switched `data_assets` on `masthead_data_product` from list to set nested attribute semantics (`Attributes Set`) to prevent spurious plan diffs caused by non-deterministic server ordering
- **Asset Normalization**: Normalized empty table string to null for `DATASET` assets to ensure stable state
- **Aggregated Diagnostics**: Plan-time validation errors are now aggregated into unified diagnostic messages instead of per-asset duplicate blocks

### Fixed

- **Empty Payload False Success**: Reject HTTP 200 responses containing empty `{"value": null}` payloads to prevent state corruption from masked upstream timeouts
- **Domain List Pagination**: Fixed domain listing to pass `limit` alongside `page` parameter to ensure all pages are properly fetched
- **Single-Attempt Cache Warm-Up**: Fixed cache warm-up to attempt bulk fetch at most once, preventing continuous redundant fallback list calls on API failures
- **Structured API Error Parsing**: Extract and surface detailed structured errors from API error responses with a fallback to raw body

## [0.2.1] - 2025-04-10

### Fixed

- Minor data products improvements and manual release workflow trigger

## [0.2.0] - 2025-04-10

### Added

- Added `masthead_data_product` resource and data source
- Added `masthead_data_domain` resource and data source

## [0.1.0] - 2025-03-20

### Added

- Initial release of `terraform-provider-masthead`
- Added `masthead_user` resource and data source

[Unreleased]: https://github.com/masthead-data/terraform-provider-masthead/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/masthead-data/terraform-provider-masthead/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/masthead-data/terraform-provider-masthead/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/masthead-data/terraform-provider-masthead/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/masthead-data/terraform-provider-masthead/releases/tag/v0.1.0
