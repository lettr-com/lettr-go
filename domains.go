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
	Domain          string     `json:"domain"`
	Status          string     `json:"status"`
	StatusLabel     string     `json:"status_label"`
	CanSend         bool       `json:"can_send"`
	CnameStatus     *string    `json:"cname_status"`
	DkimStatus      *string    `json:"dkim_status"`
	SpfStatus       *string    `json:"spf_status"`
	DmarcStatus     *string    `json:"dmarc_status"`
	TrackingDomain  *string    `json:"tracking_domain"`
	DnsProvider     *string    `json:"dns_provider"`
	IsPrimaryDomain bool       `json:"is_primary_domain"`
	DNS             *DomainDNS `json:"dns"`
	CreatedAt       string     `json:"created_at"`
	UpdatedAt       string     `json:"updated_at"`
}

// DomainDNS contains the DNS records for a domain.
type DomainDNS struct {
	DKIM  *DomainDKIM  `json:"dkim"`
	CNAME *DomainCNAME `json:"cname,omitempty"`
}

// DomainDKIM contains the DKIM DNS record details.
type DomainDKIM struct {
	Selector string `json:"selector"`
	Public   string `json:"public"`
}

// DomainCNAME contains the CNAME DNS record details.
type DomainCNAME struct {
	Host  string `json:"host"`
	Value string `json:"value"`
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

// VerifyDomainResponse is the response from verifying a domain.
type VerifyDomainResponse struct {
	Message string                 `json:"message"`
	Data    DomainVerificationView `json:"data"`
}

// DomainVerificationView contains domain verification results.
type DomainVerificationView struct {
	Domain            string                    `json:"domain"`
	DkimStatus        string                    `json:"dkim_status"`
	CnameStatus       string                    `json:"cname_status"`
	DmarcStatus       string                    `json:"dmarc_status"`
	SpfStatus         string                    `json:"spf_status"`
	IsPrimaryDomain   bool                      `json:"is_primary_domain"`
	OwnershipVerified bool                      `json:"ownership_verified"`
	Dmarc             *DmarcValidationResult    `json:"dmarc,omitempty"`
	Spf               *SpfValidationResult      `json:"spf,omitempty"`
	DNS               *DomainDnsVerificationView `json:"dns,omitempty"`
}

// DmarcValidationResult contains DMARC validation details.
type DmarcValidationResult struct {
	IsValid             bool   `json:"is_valid"`
	Status              string `json:"status"`
	FoundAtDomain       string `json:"found_at_domain,omitempty"`
	Record              string `json:"record,omitempty"`
	Policy              string `json:"policy,omitempty"`
	SubdomainPolicy     string `json:"subdomain_policy,omitempty"`
	Error               string `json:"error,omitempty"`
	CoveredByParentPolicy bool `json:"covered_by_parent_policy"`
}

// SpfValidationResult contains SPF validation details.
type SpfValidationResult struct {
	IsValid           bool   `json:"is_valid"`
	Status            string `json:"status"`
	Record            string `json:"record,omitempty"`
	Error             string `json:"error,omitempty"`
	IncludesSparkpost bool   `json:"includes_sparkpost"`
}

// DomainDnsVerificationView contains DNS verification error details.
type DomainDnsVerificationView struct {
	SpfRecord   *string `json:"spf_record,omitempty"`
	SpfError    *string `json:"spf_error,omitempty"`
	DkimRecord  *string `json:"dkim_record,omitempty"`
	DkimError   *string `json:"dkim_error,omitempty"`
	CnameRecord *string `json:"cname_record,omitempty"`
	CnameError  *string `json:"cname_error,omitempty"`
	DmarcRecord *string `json:"dmarc_record,omitempty"`
	DmarcError  *string `json:"dmarc_error,omitempty"`
}

// Verify triggers DNS record verification for a domain.
//
// Example:
//
//	result, err := client.Domains.Verify(ctx, "example.com")
func (s *DomainService) Verify(ctx context.Context, domain string) (*VerifyDomainResponse, error) {
	path := fmt.Sprintf("domains/%s/verify", url.PathEscape(domain))

	req, err := s.client.newRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}

	var resp VerifyDomainResponse
	if _, err := s.client.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
