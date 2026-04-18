# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - Unreleased

Re-synced the SDK with upstream OpenAPI spec changes ([`1e1c08a`](https://github.com/TOPOL-io/lettr/commit/1e1c08a509b7bfe8a893febac05950157ad964f8)).

### Changed

- **Breaking:** `ListEmailsResponse.Success` and `ListProjectsResponse.Success` fields removed; the API no longer returns a top-level `success` flag. Callers should check for a non-nil `error` instead.
- Bumped `Version` const to `0.3.0` (affects User-Agent header).

## [0.2.0] - 2026-04-18

Synchronized the SDK with the full Lettr OpenAPI specification.

### Added

- **Email events** endpoint: `Emails.ListEvents` (`GET /emails/events`) with filtering by event types, recipients, transmissions, bounce classes, and date range.
- **Scheduled emails**: `Emails.Schedule`, `Emails.GetScheduled`, `Emails.CancelScheduled`.
- **Domain verification**: `Domains.Verify` (`POST /domains/{domain}/verify`) with DKIM/SPF/DMARC/CNAME validation results.
- **Webhook CRUD**: `Webhooks.Create`, `Webhooks.Update`, `Webhooks.Delete`.
- **Template CRUD**: `Templates.Get`, `Templates.Update`, `Templates.Delete`, `Templates.GetMergeTags`, `Templates.GetHtml`.
- **Projects service**: new `client.Projects` with `List` method.
- New fields on `SendEmailRequest`: `Cc`, `Bcc`, `AmpHtml`, `ReplyTo`, `ReplyToName`, `Tag`, `Headers`.
- Extended `EmailEvent` with click/open event fields (`TargetLinkURL`, `TargetLinkName`, `UserAgent`, `UserAgentParsed`, `GeoIp`, `IpAddress`, `BounceClass`) and many additional CommonEventProperties fields.
- New supporting types: `UserAgentParsed`, `GeoIp`, `DomainCNAME`, `DomainVerificationView`, `DmarcValidationResult`, `SpfValidationResult`, `DomainDnsVerificationView`, `ScheduledTransmission`, `TemplateDetail`, `MergeTagChild`, `Project`.
- `DomainDetail` extended with `SpfStatus`, `DmarcStatus`, `DnsProvider`, `IsPrimaryDomain`.
- `MergeTag` extended with `Type` and `Children`.
- `CHANGELOG.md` and `RELEASING.md`.

### Changed

- **Breaking:** `Emails.Get` signature now accepts an optional `*GetEmailParams` for date-range filtering. Existing callers must pass `nil`.
- **Breaking:** `Webhook.EventTypes` changed from `[]string` to `*[]string` to distinguish "all events" (nil) from "no events" (empty slice), per spec.
- **Breaking:** `ListEmailsResponse.Data` now has an `Events` wrapper matching the spec (`Data.Events.Data`, `.TotalCount`, `.From`, `.To`, `.Pagination`). Previously decoded as empty because the struct shape did not match the API response.
- **Breaking:** `ListEmailEventsResponse.Data` now has the same `Events` wrapper for the same reason.
- **Breaking:** `GetEmailResponse.Data` is now `ScheduledTransmission` (fields `TransmissionID`, `State`, `Recipients`, `NumRecipients`, `Events[]`, etc.), matching the spec's shared transmission-detail shape. The old `GetEmailData` type has been removed.
- `ListEmailsResponse` now exposes the top-level `Success` field defined by the spec.
- **Breaking:** `EmailEvent.RcptMeta` type changed from `map[string]interface{}` to `interface{}`. The spec allows either an object (list-email items) or an array (event-stream payloads), or null; the old type failed to decode the array variant. Callers must type-assert to `map[string]interface{}` or `[]interface{}`.
- Bumped `Version` const to `0.2.0` (affects User-Agent header).

## [0.1.0] - Initial release

### Added

- Core `Client` with bearer-token authentication.
- `Emails` service: `Send`, `List`, `Get`.
- `Domains` service: `List`, `Get`, `Create`, `Delete`.
- `Webhooks` service: `List`, `Get`.
- `Templates` service: `List`, `Create`.
- `HealthCheck` and `ValidateAPIKey` on the client.
- Structured `Error` type with `IsNotFound`, `IsValidationError`, `IsUnauthorized` helpers.
