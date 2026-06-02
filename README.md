# lettr-go

The official Go SDK for the [Lettr](https://lettr.com) Email API. A typed client for emails, templates, domains, webhooks, audience, and campaigns.

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

Methods return a wrapped response (read `resp.Data`) and a structured `error`.

## Error Handling

The SDK returns typed errors you can inspect with helper predicates:

```go
resp, err := client.Emails.Send(ctx, req)
if err != nil {
	switch {
	case lettr.IsValidationError(err):
		apiErr := err.(*lettr.Error)
		for field, messages := range apiErr.Errors {
			fmt.Printf("%s: %v\n", field, messages)
		}
	case lettr.IsUnauthorized(err):
		fmt.Println("Invalid API key")
	case lettr.IsNotFound(err):
		fmt.Println("Resource not found")
	default:
		fmt.Printf("Error: %v\n", err)
	}
}
```

See [Error Handling](https://docs.lettr.com/quickstart/go/advanced#error-handling) for the full set of predicates.

## Documentation

Full guides for every service, with complete request/response details, live in the docs:

📚 **[docs.lettr.com/quickstart/go](https://docs.lettr.com/quickstart/go/quickstart)**

| Topic | Guide |
|-|-|
| Install, client, sending | [Quickstart](https://docs.lettr.com/quickstart/go/quickstart) |
| Batch sending, context & timeouts, error handling | [Advanced](https://docs.lettr.com/quickstart/go/advanced) |
| Manage Lettr templates & merge tags | [Templates](https://docs.lettr.com/quickstart/go/templates) |
| Add, verify, and manage sending domains | [Domains](https://docs.lettr.com/quickstart/go/domains) |
| Webhook endpoints for delivery & engagement events | [Webhooks](https://docs.lettr.com/quickstart/go/webhooks) |
| Lists, contacts, topics, properties, segments | [Audience](https://docs.lettr.com/quickstart/go/audience) |
| List, send, and schedule campaigns | [Campaigns](https://docs.lettr.com/quickstart/go/campaigns) |
| Endpoint reference (params & schemas) | [API Reference](https://docs.lettr.com/api-reference/introduction) |

## Versioning & Releases

- Version history: [CHANGELOG.md](./CHANGELOG.md)
- Release process: [RELEASING.md](./RELEASING.md)

This project follows [Semantic Versioning](https://semver.org/).

## License

MIT
