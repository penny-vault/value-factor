# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-05-04

### Changed
- Default benchmark from VFINX to SPY
- Upgrade pvbt dependency to v0.9.2

## [0.2.0] - 2026-05-04

### Added
- Sector cap option (`--sector-cap`, default 4) limits per-GICS-sector concentration
- Piotroski F-score screen (`--min-fscore`, default 6) drops value-trap candidates
- `Vanilla` preset disables both screens for academic baseline runs

### Changed
- Upgrade pvbt dependency to v0.9.1
- Bump strategy Version to 1.1.0

## [0.1.5] - 2026-05-03

### Changed
- Default universe is now `us-tradable` (was `SPX`)
- Upgrade pvbt dependency to v0.9.0
- Regenerate testdata snapshot

## [0.1.4] - 2026-05-01

### Changed
- Upgrade pvbt dependency to v0.8.1

## [0.1.3] - 2026-04-25

### Changed
- Upgrade pvbt dependency to v0.8.0
- Regenerate testdata snapshot for pvbt's v5 snapshot schema

## [0.1.2] - 2026-04-23

### Changed
- Upgrade pvbt dependency to v0.7.7

## [0.1.1] - 2026-04-21

### Fixed
- Remove local pvbt replace directive so the module resolves correctly outside the monorepo



## [0.1.0] - 2026-04-21

### Added
- Initial release of Value Factor strategy
- Earnings yield (E/P) factor scoring with snapshot-tested portfolio construction
- pvbt v0.7.6 support with updated API integration

[0.1.0]: https://github.com/penny-vault/value-factor/releases/tag/v0.1.0

[0.1.1]: https://github.com/penny-vault/value-factor/compare/v0.1.0...v0.1.1
[0.1.2]: https://github.com/penny-vault/value-factor/compare/v0.1.1...v0.1.2
[0.1.3]: https://github.com/penny-vault/value-factor/compare/v0.1.2...v0.1.3
[0.1.4]: https://github.com/penny-vault/value-factor/compare/v0.1.3...v0.1.4
[0.1.5]: https://github.com/penny-vault/value-factor/compare/v0.1.4...v0.1.5
[0.2.0]: https://github.com/penny-vault/value-factor/compare/v0.1.5...v0.2.0
[0.3.0]: https://github.com/penny-vault/value-factor/compare/v0.2.0...v0.3.0
