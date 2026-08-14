# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.4.0] - 2026-08-14

Covers the reworked bulk contact import (TPL-2105) and the duplicate-create fix. Everything here is additive — code written against 1.3.0 keeps compiling and sends the exact same payloads.

### Added

- **Per-contact bulk create.** `BulkCreateAudienceContactsRequest` gains a `Contacts []BulkAudienceContactRow` field: one row per contact, each with its own `Properties`, `ListIDs` and `Topics`. It is the alternative to the original flat `Emails` slice — exactly one of the two must be filled in.

  ```go
  resp, err := client.Audience.Contacts.BulkCreate(ctx, &lettr.BulkCreateAudienceContactsRequest{
      Contacts: []lettr.BulkAudienceContactRow{
          {Email: "cara@example.com", Properties: map[string]string{"plan": "pro"}},
          {Email: "dan@example.com", Topics: []lettr.AudienceTopicSubscription{
              lettr.UnsubscribeTopic("01h-promos"),
          }},
      },
      ListIDs: []string{"01h-everyone"},
  })
  ```
- New request types `BulkAudienceContactRow` and `AudienceTopicSubscription`, with the `SubscribeTopic(id)` / `UnsubscribeTopic(id)` constructors and the `TopicOptIn` / `TopicOptOut` constants.

  These say what a request should *do* with a topic, and are deliberately separate from a topic's `DefaultSubscription`, which describes how the topic behaves for a contact that says nothing. An opt-out on a topic whose default is opt-out suppresses the auto-subscription in the same request instead of needing a second call.
- **Batch-wide `ListIDs` and `Topics`,** plus `UpdateExisting`, on `BulkCreateAudienceContactsRequest`. Batch-wide lists and topics are unioned into every row; a row-level property key or opt-out wins over the batch-wide value. `UpdateExisting: true` merges properties (submitted keys overwrite, absent keys are preserved) and allows dropping a subscription. It is `omitempty`, so a legacy request stays byte-identical on the wire.
- **Bulk create now reports what happened per row.** `BulkCreateAudienceContactsData` gains `Updated`, `ErrorCount`, `Errors` (`[]BulkAudienceContactError` — `Index`, `Email`, `ErrorCode`, `Error`) and `Contacts` (`[]BulkAudienceContactRef` — `ID`, `Email`, `Created`), plus the methods `HasErrors()`, `ContactIDs()` and `IDFor(email)`. `Created` and `AlreadyExisted` keep their exact meaning, and the new fields decode to zero values against an API deployment that predates the change.

  A bulk create can **partially succeed**: rows that fail validation are skipped and returned in `Errors` while the rest of the batch commits, and the call still returns HTTP 201. A nil error does not mean every row landed — check `Data.HasErrors()`.

  Note that `AlreadyExisted` and `Updated` overlap by design. They answer different questions ("was the address already in the audience?" vs "did this request change the contact?"), so they do not sum to the row count: a contact that already existed and got attached to a list is counted in both.
- `BulkContactError*` constants for the per-row skip reasons (`missing_email`, `invalid_email`, `invalid_property_value`, `unknown_property_key`, `unknown_list`, `unknown_topic`, `invalid_topic_subscription`). `BulkAudienceContactError.ErrorCode` is a plain `string`, so a code added server-side is still readable.
- **Bulk topic subscribe/unsubscribe** — two methods on `client.Audience.Contacts`, mirroring the existing `BulkAttachToLists` / `BulkDetachFromLists` pair:
  - `BulkSubscribeToTopics(ctx, *BulkContactsTopicsRequest)` — `POST /audience/contacts/topics/bulk`, returns `Subscribed`, `AlreadySubscribed`, `TotalPairs`.
  - `BulkUnsubscribeFromTopics(ctx, *BulkContactsTopicsRequest)` — `DELETE /audience/contacts/topics/bulk` with a request body, returns `Unsubscribed`, `TotalPairs`. Pairs that do not exist are ignored.

  Both process every `ContactIDs` × `TopicIDs` combination (up to 1000 × 50). Pass `Data.ContactIDs()` from a bulk create — no ID lookup needed.
- `IsConflict(err)` and `IsContactAlreadyExists(err)` helpers, alongside the existing `IsNotFound` / `IsValidationError` / `IsUnauthorized`, plus the `ErrorCodeResourceAlreadyExists` constant.

### Changed

- Creating a contact whose email already exists now comes back as a 409 with `error_code: "resource_already_exists"`, detectable via `IsContactAlreadyExists(err)`. The API previously let this escape as HTTP 500 with the misleading `send_error` code, which names email delivery — not involved unless double opt-in is supplied. **If your retry policy retries 5xx, duplicate creates are no longer retried** — which was pointless anyway. Any error mapping or docs of yours that name `send_error` for this endpoint should be corrected.

  The returned error is still a plain `*Error`, so existing type assertions and error handling are unaffected.
- `BulkCreateAudienceContactsRequest.Emails` is now `omitempty`, so it is left out of the payload when the `Contacts` shape is used. A request that sets `Emails` serializes exactly as before.

## [1.3.0] - 2026-05-28

Adds bindings for the `/campaigns/*` resource.

### Added

- `client.Campaigns` service for the `/campaigns/*` endpoints, with methods: `List`, `Get`, `ListEvents`, `Send`, `Schedule`, and `Unschedule`. The campaigns API is read- and lifecycle-action-only (no create/update/delete); `Send`/`Schedule`/`Unschedule` require a non-sandbox key.
- Typed string constants for campaign enums: `CampaignStatus*` (`draft`, `scheduled`, `preparing`, `in_review`, `sending`, `sent`, `failed`) and `CampaignEventType*` (`injection`, `delivery`, `bounce`, `spam_complaint`, `open`, `click`, `list_unsubscribe`).
- `ListEvents` uses cursor-based pagination: `ListCampaignEventsData.NextCursor` is non-nil while more pages remain and nil once exhausted.
- `Schedule` rejects a nil `*ScheduleCampaignRequest` with a clear client-side error instead of sending the body literal `null` to the server.

### Changed

- Bumped `Version` const to `1.3.0` (affects User-Agent header).

## [1.2.0] - 2026-05-26

Adds bindings for the `/audience/*` resource and tolerance for PHP-style empty-map serialization.

### Added

- `client.Audience` service for the `/audience/*` endpoints, with nested sub-services for `Lists`, `Contacts`, `Topics`, `Properties`, and `Segments`. Covers CRUD, bulk operations, and contact↔list/topic attachments.
- Typed string constants for audience enums: `ContactStatus*`, `TopicDefaultSubscription*`, `TopicVisibility*`, `PropertyType*`, and `SegmentOperator*`.
- `ContactProperties` named map type for `AudienceContact.Properties`, with a tolerant `UnmarshalJSON` that accepts the PHP/Laravel empty-associative-array form (`"properties": []`) as well as the canonical object form (`"properties": {...}`).
- Tolerant `UnmarshalJSON` on `Error`: a 422 response carrying `"errors": []` (or `null`) now decodes to an empty `Errors` map instead of failing with a confusing `cannot unmarshal array into Go struct field` decode error. The `Errors` field type is unchanged.
- `NullString` tri-state helper for nullable PATCH fields (`UpdateAudienceTopicRequest.Description`, `UpdateAudiencePropertyRequest.FallbackValue`, `UpdateAudienceSegmentRequest.ListID`). Field names and zero-value semantics mirror `database/sql.NullString`: nil pointer omits the field, `&NullString{}` (zero value) sends JSON null, and `NewNullString("x")` (or `&NullString{Valid: true, String: "x"}`) sends the string.

### Changed

- Bumped `Version` const to `1.2.0` (affects User-Agent header).

## [1.1.0] - Unreleased

Sync with the updated webhook contract.

### Added

- `URL` field on `UpdateWebhookRequest` matching the new `PUT /webhooks/{id}` schema.
- Exported event-type constants (`EventMessageDelivery`, `EventEngagementClick`, `EventUnsubscribeList`, …) covering the namespaced webhook event names.

### Changed

- Webhook event strings are now namespaced (`message.delivery`, `engagement.click`, …) on both request and response sides. The `[]string` field types are unchanged; callers must update string values they pass.
- Bumped `Version` const to `1.1.0` (affects User-Agent header).

### Deprecated

- `UpdateWebhookRequest.Target` — use `URL` instead. The field continues to serialize as `target` for now and will be removed in a future major release.

## [1.0.0] - 2026-04-20

First stable release. The public API is now committed to Semantic Versioning — breaking changes will require a new major version. No code changes from [0.3.0].

### Changed

- Bumped `Version` const to `1.0.0` (affects User-Agent header).

## [0.3.0] - 2026-04-20

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
