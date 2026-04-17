# lettr-go

The official Go SDK for the [Lettr](https://lettr.com) Email API for Artisans. Send transactional emails with tracking, attachments, and template support.

## Installation

```bash
go get github.com/lettr-com/lettr-go
```

Requires Go 1.21 or later.

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	lettr "github.com/lettr-com/lettr-go"
)

func main() {
	client := lettr.NewClient("your-api-key")

	resp, err := client.Emails.Send(context.Background(), &lettr.SendEmailRequest{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Hello from Lettr",
		Html:    "<h1>Hello!</h1>",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Email sent! Request ID: %s\n", resp.Data.RequestID)
}
```

## Usage

### Create a Client

```go
// Default client (30s timeout)
client := lettr.NewClient("your-api-key")

// Custom HTTP client
client := lettr.NewClientWithHTTPClient("your-api-key", &http.Client{
    Timeout: 60 * time.Second,
})
```

### Send an Email

```go
resp, err := client.Emails.Send(ctx, &lettr.SendEmailRequest{
    From:     "sender@example.com",
    FromName: "Sender Name",
    To:       []string{"recipient@example.com"},
    Cc:       []string{"cc@example.com"},
    Bcc:      []string{"bcc@example.com"},
    Subject:  "Welcome to Lettr",
    Html:     "<h1>Welcome!</h1>",
    Text:     "Welcome!",
    ReplyTo:  "reply@example.com",
    Tag:      "welcome",
    Headers: map[string]string{
        "X-Custom-Header": "value",
    },
    Options: &lettr.SendEmailOptions{
        ClickTracking: boolPtr(true),
        OpenTracking:  boolPtr(true),
    },
})
```

### Send with Template

```go
resp, err := client.Emails.Send(ctx, &lettr.SendEmailRequest{
    From:         "hello@example.com",
    To:           []string{"john@example.com"},
    Subject:      "Welcome, {{first_name}}!",
    TemplateSlug: "welcome-email",
    SubstitutionData: map[string]interface{}{
        "first_name": "John",
        "company":    "Acme Inc",
    },
})
```

### Send with Attachments

```go
resp, err := client.Emails.Send(ctx, &lettr.SendEmailRequest{
    From:    "billing@example.com",
    To:      []string{"customer@example.com"},
    Subject: "Your Invoice",
    Html:    "<p>Please find your invoice attached.</p>",
    Attachments: []lettr.Attachment{
        {
            Name: "invoice.pdf",
            Type: "application/pdf",
            Data: base64EncodedContent,
        },
    },
})
```

### List Sent Emails

```go
emails, err := client.Emails.List(ctx, &lettr.ListEmailsParams{
    PerPage:    10,
    Recipients: "user@example.com",
    From:       "2024-01-01",
    To:         "2024-01-31",
})

for _, email := range emails.Data.Events.Data {
    fmt.Printf("%s -> %s: %s\n", email.FriendlyFrom, email.RcptTo, email.Subject)
}

// Paginate with cursor
if emails.Data.Events.Pagination.NextCursor != nil {
    nextPage, err := client.Emails.List(ctx, &lettr.ListEmailsParams{
        Cursor: *emails.Data.Events.Pagination.NextCursor,
    })
}
```

### Get Email Details

```go
details, err := client.Emails.Get(ctx, "request-id-from-send", nil)

// With optional date range filter
details, err := client.Emails.Get(ctx, "request-id-from-send", &lettr.GetEmailParams{
    From: "2024-01-01",
    To:   "2024-01-31",
})

fmt.Printf("Transmission %s state=%s\n", details.Data.TransmissionID, details.Data.State)
for _, event := range details.Data.Events {
    fmt.Printf("[%s] %s at %s\n", event.Type, event.RcptTo, event.Timestamp)
}
```

### List Email Events

```go
events, err := client.Emails.ListEvents(ctx, &lettr.ListEmailEventsParams{
    Events:     []string{"delivery", "bounce", "open", "click"},
    Recipients: "user@example.com",
    PerPage:    50,
    From:       "2024-01-01",
    To:         "2024-01-31",
})

for _, event := range events.Data.Events.Data {
    fmt.Printf("[%s] %s -> %s at %s\n", event.Type, event.FriendlyFrom, event.RcptTo, event.Timestamp)
}
```

### Schedule an Email

```go
// Schedule for future delivery (min 5 minutes, max 3 days ahead)
resp, err := client.Emails.Schedule(ctx, &lettr.ScheduleEmailRequest{
    SendEmailRequest: lettr.SendEmailRequest{
        From:    "sender@example.com",
        To:      []string{"recipient@example.com"},
        Subject: "Scheduled Hello",
        Html:    "<h1>Hello!</h1>",
    },
    ScheduledAt: "2024-12-25T10:00:00Z",
})
fmt.Printf("Transmission ID: %s\n", resp.Data.TransmissionID)

// Check scheduled email status
scheduled, err := client.Emails.GetScheduled(ctx, resp.Data.TransmissionID)
fmt.Printf("Status: %s\n", scheduled.Data.Status)

// Cancel a scheduled email
err = client.Emails.CancelScheduled(ctx, resp.Data.TransmissionID)
```

### Domains

```go
// List all domains
domains, err := client.Domains.List(ctx)

// Get domain details (including DNS records)
domain, err := client.Domains.Get(ctx, "example.com")

// Register a new domain
created, err := client.Domains.Create(ctx, &lettr.CreateDomainRequest{
    Domain: "example.com",
})

// Verify domain DNS records
verification, err := client.Domains.Verify(ctx, "example.com")
fmt.Printf("DKIM: %s, SPF: %s, DMARC: %s\n",
    verification.Data.DkimStatus,
    verification.Data.SpfStatus,
    verification.Data.DmarcStatus,
)

// Delete a domain
err = client.Domains.Delete(ctx, "example.com")
```

### Webhooks

```go
// List all webhooks
webhooks, err := client.Webhooks.List(ctx)

// Get webhook details
webhook, err := client.Webhooks.Get(ctx, "webhook-id")

// Create a webhook (receive all events)
created, err := client.Webhooks.Create(ctx, &lettr.CreateWebhookRequest{
    Name:       "My Webhook",
    URL:        "https://example.com/webhook",
    AuthType:   "none",
    EventsMode: "all",
})

// Create a webhook (receive selected events with basic auth)
created, err = client.Webhooks.Create(ctx, &lettr.CreateWebhookRequest{
    Name:         "Delivery Webhook",
    URL:          "https://example.com/webhook",
    AuthType:     "basic",
    AuthUsername: "user",
    AuthPassword: "pass",
    EventsMode:   "selected",
    Events:       []string{"delivery", "bounce", "spam_complaint"},
})

// Update a webhook
active := false
updated, err := client.Webhooks.Update(ctx, "webhook-id", &lettr.UpdateWebhookRequest{
    Name:   "Renamed Webhook",
    Active: &active,
})

// Delete a webhook
err = client.Webhooks.Delete(ctx, "webhook-id")
```

### Templates

```go
// List templates
templates, err := client.Templates.List(ctx, &lettr.ListTemplatesParams{
    ProjectID: 5,
    PerPage:   10,
    Page:      1,
})

// Get template details
template, err := client.Templates.Get(ctx, "welcome-email", nil)
// With specific project
template, err = client.Templates.Get(ctx, "welcome-email", &lettr.GetTemplateParams{
    ProjectID: 5,
})

// Create a template
created, err := client.Templates.Create(ctx, &lettr.CreateTemplateRequest{
    Name: "Welcome Email",
    Html: "<h1>Hello {{FIRST_NAME}}!</h1>",
})

// Update a template (creates a new active version)
updated, err := client.Templates.Update(ctx, "welcome-email", &lettr.UpdateTemplateRequest{
    Html: "<h1>Updated Hello {{FIRST_NAME}}!</h1>",
})

// Get merge tags for a template
tags, err := client.Templates.GetMergeTags(ctx, "welcome-email", nil)
for _, tag := range tags.Data.MergeTags {
    fmt.Printf("Tag: %s (required: %v)\n", tag.Key, tag.Required)
}

// Get rendered HTML content
html, err := client.Templates.GetHtml(ctx, &lettr.GetTemplateHtmlParams{
    ProjectID: 5,
    Slug:      "welcome-email",
})
fmt.Println(html.Data.Html)

// Delete a template
err = client.Templates.Delete(ctx, "welcome-email", nil)
```

### Projects

```go
// List projects
projects, err := client.Projects.List(ctx, nil)

for _, project := range projects.Data.Projects {
    fmt.Printf("Project: %s (ID: %d)\n", project.Name, project.ID)
}

// With pagination
projects, err = client.Projects.List(ctx, &lettr.ListProjectsParams{
    PerPage: 10,
    Page:    2,
})
```

### System

```go
// Health check (no auth required)
health, err := client.HealthCheck(ctx)

// Validate API key
auth, err := client.ValidateAPIKey(ctx)
fmt.Printf("Team ID: %d\n", auth.Data.TeamID)
```

## Error Handling

The SDK returns structured errors with HTTP status codes and API error codes:

```go
resp, err := client.Emails.Send(ctx, &lettr.SendEmailRequest{})
if err != nil {
    // Check specific error types
    if lettr.IsValidationError(err) {
        apiErr := err.(*lettr.Error)
        for field, messages := range apiErr.Errors {
            fmt.Printf("%s: %v\n", field, messages)
        }
    } else if lettr.IsUnauthorized(err) {
        fmt.Println("Invalid API key")
    } else if lettr.IsNotFound(err) {
        fmt.Println("Resource not found")
    } else {
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Available Services

| Service | Methods |
|---------|---------|
| `client.Emails` | `Send`, `List`, `Get`, `ListEvents`, `Schedule`, `GetScheduled`, `CancelScheduled` |
| `client.Domains` | `List`, `Get`, `Create`, `Delete`, `Verify` |
| `client.Webhooks` | `List`, `Get`, `Create`, `Update`, `Delete` |
| `client.Templates` | `List`, `Get`, `Create`, `Update`, `Delete`, `GetMergeTags`, `GetHtml` |
| `client.Projects` | `List` |
| `client` (system) | `HealthCheck`, `ValidateAPIKey` |

## Versioning & Releases

- Version history: [CHANGELOG.md](./CHANGELOG.md)
- Release process: [RELEASING.md](./RELEASING.md)

This project follows [Semantic Versioning](https://semver.org/).

## License

MIT
