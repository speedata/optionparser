# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- `GenerateCompletion(shell, programName, w)` generates shell completion scripts for bash, zsh and fish from the registered options and commands.
- Panic recovery in user callback functions to prevent crashes from user code errors.

### Fixed
- Parser now catches panics in user callback functions (`func(string)` and `func()`) and returns them as errors instead of crashing the program (fixes #8).

---

## [v1.1.0] - 2025-10-09
### Changed
- `Help()` now writes to an `io.Writer` (defaults to `os.Stdout`) instead of hardcoding standard output.
- `Parse()` and `ParseFrom()` now return `ErrHelp` instead of calling `os.Exit(0)` when `--help` is requested.
- Reworked internal formatting to avoid potential panics with small `Start` or `Stop` values.

### Fixed
- Restored support for short negations like `-no-c`.
- Corrected regular expressions for option detection (`--long`, `-x`).
- Added bounds checks in `splitOn()` to prevent index out of range errors.
- Improved handling of `stringslice` and `stringmap` assignments (no empty entries).
- Improved word wrapping and help text formatting for better alignment and safety.
- Prevented duplicate registration of short or long options.
- Maintained backward compatibility with all existing APIs.

### Developer Experience
- Added internal error `ErrHelp` for easier integration in CLI tools and libraries.
- The parser can now be tested without console output by setting `OptionParser.Out`.

---


[Unreleased]: https://github.com/speedata/optionparser/compare/v1.1.0...HEAD
[v1.1.0]: https://github.com/speedata/optionparser/compare/v1.0.5...v1.1.0
[v1.0.5]: https://github.com/speedata/optionparser/releases/tag/v1.0.5