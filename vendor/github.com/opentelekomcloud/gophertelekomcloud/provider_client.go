package golangsdk

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/opentelekomcloud/gophertelekomcloud/internal/extract"
)

// DefaultUserAgent is the default User-Agent string set in the request header.
const DefaultUserAgent = "golangsdk/2.0.0"

// UserAgent represents a User-Agent header.
type UserAgent struct {
	// prepend is the slice of User-Agent strings to prepend to DefaultUserAgent.
	// All the strings to prepend are accumulated and prepended in the Join method.
	prepend []string
}

// Prepend prepends a user-defined string to the default User-Agent string. Users
// may pass in one or more strings to prepend.
func (ua *UserAgent) Prepend(s ...string) {
	ua.prepend = append(s, ua.prepend...)
}

// Join concatenates all the user-defined User-Agent strings with the default
// Gophercloud User-Agent string.
func (ua *UserAgent) Join() string {
	uaSlice := append(ua.prepend, DefaultUserAgent)
	return strings.Join(uaSlice, " ")
}

// ProviderClient stores details that are required to interact with any
// services within a specific provider's API.
//
// Generally, you acquire a ProviderClient by calling the NewClient method in
// the appropriate provider's child package, providing whatever authentication
// credentials are required.
type ProviderClient struct {
	// IdentityBase is the base URL used for a particular provider's identity
	// service - it will be used when issuing authentication requests. It
	// should point to the root resource of the identity service, not a specific
	// identity version.
	IdentityBase string

	// IdentityEndpoint is the identity endpoint. This may be a specific version
	// of the identity service. If this is the case, this endpoint is used rather
	// than querying versions first.
	IdentityEndpoint string

	// TokenID is the ID of the most recently issued valid token.
	// NOTE: Aside from within a custom ReauthFunc, this field shouldn't be set by an application.
	// To safely read or write this value, call `Token` or `SetToken`, respectively
	TokenID string

	// ProjectID is the ID of project to which User is authorized.
	ProjectID string

	// UserID is the ID of the authorized user
	UserID string

	// DomainID is the ID of project to which User is authorized.
	DomainID string

	// RegionID is the Name of region to which User is authorized.
	RegionID string

	// EndpointLocator describes how this provider discovers the endpoints for
	// its constituent services.
	EndpointLocator EndpointLocator

	// HTTPClient allows users to interject arbitrary http, https, or other transit behaviors.
	HTTPClient http.Client

	// UserAgent represents the User-Agent header in the HTTP request.
	UserAgent UserAgent

	// MaxBackoffRetries set the maximum number of backoffs. When not set, defaults to defaultMaxBackoffRetryLimit
	MaxBackoffRetries *int
	// BackoffRetryTimeout specifies time before next retry on 429. When not set, defaults to defaultBackoffTimeout
	BackoffRetryTimeout *time.Duration
	// ReauthFunc is the function used to re-authenticate the user if the request
	// fails with a 401 HTTP response code. This a needed because there may be multiple
	// authentication functions for different Identity service versions.
	ReauthFunc func() error

	// AKSKAuthOptions provides the value for AK/SK authentication, it should be nil if you use token authentication,
	// Otherwise, it must have a value
	AKSKAuthOptions AKSKAuthOptions

	mut *sync.RWMutex

	reauthmut *reauthlock
}

type reauthlock struct {
	sync.RWMutex
	reauthing bool
}

// AuthenticatedHeaders returns a map of HTTP headers that are common for all
// authenticated service requests.
func (client *ProviderClient) AuthenticatedHeaders() (m map[string]string) {
	if client.reauthmut != nil {
		client.reauthmut.RLock()
		if client.reauthmut.reauthing {
			client.reauthmut.RUnlock()
			return
		}
		client.reauthmut.RUnlock()
	}
	t := client.Token()
	if t == "" {
		return
	}
	return map[string]string{"X-Auth-Token": t}
}

// UseTokenLock creates a mutex that is used to allow safe concurrent access to the auth token.
// If the application's ProviderClient is not used concurrently, this doesn't need to be called.
func (client *ProviderClient) UseTokenLock() {
	client.mut = new(sync.RWMutex)
	client.reauthmut = new(reauthlock)
}

// Token safely reads the value of the auth token from the ProviderClient. Applications should
// call this method to access the token instead of the TokenID field
func (client *ProviderClient) Token() string {
	if client.mut != nil {
		client.mut.RLock()
		defer client.mut.RUnlock()
	}
	return client.TokenID
}

// SetToken safely sets the value of the auth token in the ProviderClient. Applications may
// use this method in a custom ReauthFunc
func (client *ProviderClient) SetToken(t string) {
	if client.mut != nil {
		client.mut.Lock()
		defer client.mut.Unlock()
	}
	client.TokenID = t
}

// RequestOpts customizes the behavior of the provider.Request() method.
type RequestOpts struct {
	// JSONBody, if provided, will be encoded as JSON and used as the body of the HTTP request. The
	// content type of the request will default to "application/json" unless overridden by MoreHeaders.
	// It's an error to specify both a JSONBody and a RawBody.
	JSONBody interface{}
	// RawBody contains an io.Reader that will be consumed by the request directly. No content-type
	// will be set unless one is provided explicitly by MoreHeaders.
	RawBody io.Reader
	// JSONResponse, if provided, will be populated with the contents of the response body parsed as JSON.
	// Not that setting it will drain and close Response.Body.
	// Deprecated: Use http.Response Body instead.
	JSONResponse interface{}
	// OkCodes contains a list of numeric HTTP status codes that should be interpreted as success. If
	// the response has a different code, an error will be returned.
	OkCodes []int
	// MoreHeaders specifies additional HTTP headers to be provide on the request. If a header is
	// provided with a blank value (""), that header will be *omitted* instead: use this to suppress
	// the default Accept header or an inferred Content-Type, for example.
	MoreHeaders map[string]string
	// ErrorContext specifies the resource error type to return if an error is encountered.
	// This lets resources override default error messages based on the response status code.
	ErrorContext error

	// RetryCount specifies number of times retriable errors (502, 504) will be retried
	RetryCount *int
	// RetryTimeout specifies time before next retry
	RetryTimeout *time.Duration
}

var applicationJSON = "application/json"

// Request performs an HTTP request using the ProviderClient's current HTTPClient. An authentication
// header will automatically be provided.
func (client *ProviderClient) Request(method, url string, options *RequestOpts) (*http.Response, error) {
	var body io.Reader
	var contentType *string

	// Derive the content body by either encoding an arbitrary object as JSON, or by taking a provided
	// io.ReadSeeker as-is. Default the content-type to application/json.
	if options.JSONBody != nil {
		if options.RawBody != nil {
			panic("Please provide only one of JSONBody or RawBody to golangsdk.Request().")
		}

		rendered, err := extract.JsonMarshal(options.JSONBody)
		if err != nil {
			return nil, err
		}

		body = bytes.NewReader(rendered)
		contentType = &applicationJSON
	}

	if options.RawBody != nil {
		body = options.RawBody
	}

	// Construct the http.Request.
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// Populate the request headers. Apply options.MoreHeaders last, to give the caller the chance to
	// modify or omit any header.
	if contentType != nil {
		req.Header.Set("Content-Type", *contentType)
	}
	req.Header.Set("Accept", applicationJSON)

	// Set the User-Agent header
	req.Header.Set("User-Agent", client.UserAgent.Join())

	if options.MoreHeaders != nil {
		for k, v := range options.MoreHeaders {
			if v != "" {
				req.Header.Set(k, v)
			} else {
				req.Header.Del(k)
			}
		}
	}

	// get the latest token from client
	for k, v := range client.AuthenticatedHeaders() {
		req.Header.Set(k, v)
	}

	// Set connection parameter to close the connection immediately when we've got the response
	req.Close = true

	prereqtok := req.Header.Get("X-Auth-Token")

	if client.AKSKAuthOptions.AccessKey != "" {
		Sign(req, SignOptions{
			AccessKey: client.AKSKAuthOptions.AccessKey,
			SecretKey: client.AKSKAuthOptions.SecretKey,
		})
		if client.AKSKAuthOptions.ProjectId != "" && client.AKSKAuthOptions.DomainID == "" {
			req.Header.Set("X-Project-Id", client.AKSKAuthOptions.ProjectId)
		}
		if client.AKSKAuthOptions.DomainID != "" {
			req.Header.Set("X-Domain-Id", client.AKSKAuthOptions.DomainID)
		}
		if client.AKSKAuthOptions.SecurityToken != "" {
			req.Header.Set("X-Security-Token", client.AKSKAuthOptions.SecurityToken)
		}
	}

	// Issue the request.
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Allow default OkCodes if none explicitly set
	if options.OkCodes == nil {
		options.OkCodes = defaultOkCodes(method)
	}

	if options.RetryCount == nil {
		defaultRetryLimit := 1
		options.RetryCount = &defaultRetryLimit
	}

	if options.RetryTimeout == nil {
		defaultRetryTimeout := 500 * time.Millisecond
		options.RetryTimeout = &defaultRetryTimeout
	}

	if client.MaxBackoffRetries == nil {
		defaultMaxBackoffRetryLimit := 20
		client.MaxBackoffRetries = &defaultMaxBackoffRetryLimit
	}

	if client.BackoffRetryTimeout == nil {
		defaultBackoffTimeout := 60 * time.Second
		client.BackoffRetryTimeout = &defaultBackoffTimeout
	}

	// Validate the HTTP response status.
	var ok bool
	for _, code := range options.OkCodes {
		if resp.StatusCode == code {
			ok = true
			break
		}
	}

	if !ok {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		respErr := ErrUnexpectedResponseCode{
			URL:      url,
			Method:   method,
			Expected: options.OkCodes,
			Actual:   resp.StatusCode,
			Body:     body,
		}

		errType := options.ErrorContext
		switch resp.StatusCode {
		case http.StatusBadRequest:
			err = ErrDefault400{respErr}
			if error400er, ok := errType.(Err400er); ok {
				err = error400er.Error400(respErr)
			}
		case http.StatusUnauthorized:
			if client.ReauthFunc != nil {
				if client.mut != nil {
					client.mut.Lock()
					client.reauthmut.Lock()
					client.reauthmut.reauthing = true
					client.reauthmut.Unlock()
					if curtok := client.TokenID; curtok == prereqtok {
						err = client.ReauthFunc()
					}
					client.reauthmut.Lock()
					client.reauthmut.reauthing = false
					client.reauthmut.Unlock()
					client.mut.Unlock()
				} else {
					err = client.ReauthFunc()
				}
				if err != nil {
					e := &ErrUnableToReauthenticate{}
					e.ErrOriginal = respErr
					return nil, e
				}
				if options.RawBody != nil {
					if seeker, ok := options.RawBody.(io.Seeker); ok {
						_, e := seeker.Seek(0, 0)
						if e != nil {
							return nil, e
						}
					}
				}
				resp, err = client.Request(method, url, options)
				if err != nil {
					e := &ErrErrorAfterReauthentication{}
					e.ErrOriginal = err
					return nil, e
				}
				return resp, nil
			}
			err = ErrDefault401{respErr}
			if error401er, ok := errType.(Err401er); ok {
				err = error401er.Error401(respErr)
			}
		case http.StatusForbidden:
			err = ErrDefault403{respErr}
			if error403er, ok := errType.(Err403er); ok {
				err = error403er.Error403(respErr)
			}
		case http.StatusNotFound:
			err = ErrDefault404{respErr}
			if error404er, ok := errType.(Err404er); ok {
				err = error404er.Error404(respErr)
			}
		case http.StatusMethodNotAllowed:
			err = ErrDefault405{respErr}
			if error405er, ok := errType.(Err405er); ok {
				err = error405er.Error405(respErr)
			}
		case http.StatusRequestTimeout:
			err = ErrDefault408{respErr}
			if error408er, ok := errType.(Err408er); ok {
				err = error408er.Error408(respErr)
			}
		case http.StatusConflict:
			err = ErrDefault409{respErr}
			if error409er, ok := errType.(Err409er); ok {
				err = error409er.Error409(respErr)
			}
		case http.StatusTooManyRequests:
			err = ErrDefault429{respErr}
			if error429er, ok := errType.(Err429er); ok {
				err = error429er.Error429(respErr)
			}
			if *client.MaxBackoffRetries > 0 {
				*client.MaxBackoffRetries -= 1
				time.Sleep(*client.BackoffRetryTimeout)
				return client.Request(method, url, options)
			}
		case http.StatusInternalServerError:
			err = ErrDefault500{respErr}
			if error500er, ok := errType.(Err500er); ok {
				err = error500er.Error500(respErr)
			}
		case http.StatusBadGateway, http.StatusGatewayTimeout: // gateway errors
			if *options.RetryCount > 0 {
				*options.RetryCount -= 1
				time.Sleep(*options.RetryTimeout)
				return client.Request(method, url, options)
			}
		case http.StatusServiceUnavailable:
			err = ErrDefault503{respErr}
			if error503er, ok := errType.(Err503er); ok {
				err = error503er.Error503(respErr)
			}
		}

		if err == nil {
			err = respErr
		}

		return resp, err
	}

	// Parse the response body as JSON, if requested to do so.
	// TODO: When all refactoring of the extract is done, remove this.
	if options.JSONResponse != nil && resp.StatusCode != http.StatusNoContent {
		switch r := options.JSONResponse.(type) {
		case *[]byte:
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("error reading response body: %w", err)
			}
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					_ = fmt.Errorf("error in closing : %w", err)
				}
			}(resp.Body)

			*r = data
		default:
			if err := extract.Into(resp.Body, &r); err != nil {
				return nil, err
			}
		}
	}

	return resp, nil
}

func defaultOkCodes(method string) []int {
	switch method {
	case "GET":
		return []int{200}
	case "POST":
		return []int{201, 202}
	case "PUT":
		return []int{201, 202}
	case "PATCH":
		return []int{200, 204}
	case "DELETE":
		return []int{202, 204}
	case "HEAD":
		return []int{204, 206}
	}
	return []int{}
}
