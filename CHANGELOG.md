# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.5] - 2023-04-06

### Added

- Priority queue for handling external api requests 
- Indexation for higher level

### Fixed

- Postgres function flow, made them more universal 

## [1.0.4] - 2023-03-30

### Added

- Registration will be every `n` minutes to be sure that module is registered
- Registration logic

### Changed

- Add submodule to get user by id request

## [1.0.3] - 2023-03-27

### Added

- Worker delete users and permissions that wasn't updated for some time
- `updated_at` column in users table

## [1.0.2] - 2023-03-22

### Added

- Request to send roles by user status

## [1.0.1] - 2023-03-17

### Added

- Handling `Too Many requests` error from `API`

## [1.0.0] - 2023-03-15

### Added

- Sender.
- Receiver.
- Database.
- API handlers.

[1.0.0]: https://gitlab.com/distributed_lab/acs/github-module/-/tree/feature/review_fixes
[1.0.1]: https://gitlab.com/distributed_lab/acs/gitlab-module/-/compare/feature/review_fixes...feature/many_requests_handle
[1.0.2]: https://gitlab.com/distributed_lab/acs/gitlab-module/-/compare/feature/review_fixes...feature/many_requests_handle