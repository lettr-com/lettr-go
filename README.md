# lettr-go

The official Go SDK for the [Lettr](https://lettr.com) Email API for Artisans. Send transactional emails with tracking, attachments, and template support.

## Installation

```bash
go get github.com/lettr/lettr-go
```

Requires Go 1.21 or later.

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	lettr "github.com/lettr/lettr-go"
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
    Subject:  "Welcome to Lettr",
    Html:     "<h1>Welcome!</h1>",
    Text:     "Welcome!",
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

for _, email := range emails.Data.Results {
    fmt.Printf("%s -> %s: %s\n", email.FriendlyFrom, email.RcptTo, email.Subject)
}

// Paginate with cursor
if emails.Data.Pagination.NextCursor != nil {
    nextPage, err := client.Emails.List(ctx, &lettr.ListEmailsParams{
        Cursor: *emails.Data.Pagination.NextCursor,
    })
}
```

### Get Email Details

```go
details, err := client.Emails.Get(ctx, "request-id-from-send")

for _, event := range details.Data.Results {
    fmt.Printf("[%s] %s at %s\n", event.Type, event.RcptTo, event.Timestamp)
}
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

// Delete a domain
err := client.Domains.Delete(ctx, "example.com")
```

### Webhooks

```go
// List all webhooks
webhooks, err := client.Webhooks.List(ctx)

// Get webhook details
webhook, err := client.Webhooks.Get(ctx, "webhook-id")
```

### Templates

```go
// List templates
templates, err := client.Templates.List(ctx, &lettr.ListTemplatesParams{
    ProjectID: 5,
    PerPage:   10,
    Page:      1,
})

// Create a template
template, err := client.Templates.Create(ctx, &lettr.CreateTemplateRequest{
    Name: "Welcome Email",
    Html: "<h1>Hello {{FIRST_NAME}}!</h1>",
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

## License

MIT
