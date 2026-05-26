package lettr

// AudienceService groups all audience-related operations under one namespace.
// The endpoint methods live on its sub-services: Lists, Contacts, Topics,
// Properties, and Segments.
//
// Example:
//
//	lists, err := client.Audience.Lists.List(ctx, nil)
//	contact, err := client.Audience.Contacts.Get(ctx, "0193e6b0-9c1d-7d4f-a8f1-cef9a1b2d3e4")
type AudienceService struct {
	Lists      *AudienceListService
	Contacts   *AudienceContactService
	Topics     *AudienceTopicService
	Properties *AudiencePropertyService
	Segments   *AudienceSegmentService
}

func newAudienceService(c *Client) *AudienceService {
	return &AudienceService{
		Lists:      &AudienceListService{client: c},
		Contacts:   &AudienceContactService{client: c},
		Topics:     &AudienceTopicService{client: c},
		Properties: &AudiencePropertyService{client: c},
		Segments:   &AudienceSegmentService{client: c},
	}
}
