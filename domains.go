package lettr

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// DomainService handles communication with the domain-related endpoints
// of the Lettr API.
type DomainService struct {
	client *Client
}

// Domain represents a sending domain.
type Domain struct {
	Domain      string  `json:"domain"`
	Status      string  `json:"status"`
	StatusLabel string  `json:"status_label"`
	CanSend     bool    `json:"can_send"`
	CnameStatus *string `json:"cname_status"`
	DkimStatus  *string `json:"dkim_status"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// DomainDetail represents detailed information about a sending domain,
// including DNS records and tracking domain configuration.
type DomainDetail struct {
	Domain         string     `json:"domain"`
	Status         string     `json:"status"`
	StatusLabel    string     `json:"status_label"`
	CanSend        bool       `json:"can_send"`
	CnameStatus    *string    `json:"cname_status"`
	DkimStatus     *string    `json:"dkim_status"`
	TrackingDomain *string    `json:"tracking_domain"`
	DNS            *DomainDNS `json:"dns"`
	CreatedAt      string     `json:"created_at"`
	UpdatedAt      string     `json:"updated_at"`
}

// DomainDNS contains the DNS records for a domain.
type DomainDNS struct {
	DKIM *DomainDKIM `json:"dkim"`
}

// DomainDKIM contains the DKIM DNS record details.
type DomainDKIM struct {
	Selector string `json:"selector"`
	Public   string `json:"public"`
}

// CreateDomainRequest represents the request body for creating a domain.
type CreateDomainRequest struct {
	// Domain is the domain name to register (e.g. "example.com").
	Domain string `json:"domain"`
}

// ListDomainsResponse is the response from listing domains.
type ListDomainsResponse struct {
	Message string          `json:"message"`
	Data    ListDomainsData `json:"data"`
}

// ListDomainsData contains the list of domains.
type ListDomainsData struct {
	Domains []Domain `json:"domains"`
}

// GetDomainResponse is the response from getting a single domain.
type GetDomainResponse struct {
	Message string       `json:"message"`
	Data    DomainDetail `json:"data"`
}

// CreateDomainResponse is the response from creating a domain.
type CreateDomainResponse struct {
	Message string           `json:"message"`
	Data    CreateDomainData `json:"data"`
}

// CreateDomainData contains the result of creating a domain.
type CreateDomainData struct {
	Domain      string      `json:"domain"`
	Status      string      `json:"status"`
	StatusLabel string      `json:"status_label"`
	DKIM        *DomainDKIM `json:"dkim"`
}

// List retrieves all sending domains registered with your account.
//
// Example:
//
//	domains, err := client.Domains.List(ctx)
func (s *DomainService) List(ctx context.Context) (*ListDomainsResponse, error) {
	req, err := s.client.newRequest(ctx, http.MethodGet, "domains", nil)
	if err != nil {
		return nil, err
	}

	var resp ListDomainsResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get retrieves details of a single sending domain including DNS records.
//
// Example:
//
//	domain, err := client.Domains.Get(ctx, "example.com")
func (s *DomainService) Get(ctx context.Context, domain string) (*GetDomainResponse, error) {
	path := fmt.Sprintf("domains/%s", url.PathEscape(domain))

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp GetDomainResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Create registers a new sending domain with your account.
// The domain will start in a pending state until verified.
//
// Example:
//
//	created, err := client.Domains.Create(ctx, &lettr.CreateDomainRequest{
//	    Domain: "example.com",
//	})
func (s *DomainService) Create(ctx context.Context, params *CreateDomainRequest) (*CreateDomainResponse, error) {
	req, err := s.client.newRequest(ctx, http.MethodPost, "domains", params)
	if err != nil {
		return nil, err
	}

	var resp CreateDomainResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Delete removes a sending domain. The domain will no longer be available
// for sending emails.
//
// Example:
//
//	err := client.Domains.Delete(ctx, "example.com")
func (s *DomainService) Delete(ctx context.Context, domain string) error {
	path := fmt.Sprintf("domains/%s", url.PathEscape(domain))

	req, err := s.client.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(req, nil)
	return err
}
