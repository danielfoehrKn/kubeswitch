package openstack

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/identity/v3/catalog"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/identity/v3/domains"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/identity/v3/projects"
	tokens3 "github.com/opentelekomcloud/gophertelekomcloud/openstack/identity/v3/tokens"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/utils"
	"github.com/opentelekomcloud/gophertelekomcloud/pagination"
)

const (
	// v3 represents Keystone v3.
	// The version can be anything from v3 to v3.x.
	v3 = "v3"
)

/*
NewClient prepares an unauthenticated ProviderClient instance.
Most users will probably prefer using the AuthenticatedClient function
instead.

This is useful if you wish to explicitly control the version of the identity
service that's used for authentication explicitly, for example.

A basic example of using this would be:

	ao, err := openstack.AuthOptionsFromEnv()
	provider, err := openstack.NewClient(ao.IdentityEndpoint)
	client, err := openstack.NewIdentityV3(provider, golangsdk.EndpointOpts{})
*/
func NewClient(endpoint string) (*golangsdk.ProviderClient, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	u.RawQuery, u.Fragment = "", ""

	var base string
	versionRe := regexp.MustCompile("v[0-9.]+/?")
	if version := versionRe.FindString(u.Path); version != "" {
		base = strings.Replace(u.String(), version, "", -1)
	} else {
		base = u.String()
	}

	endpoint = golangsdk.NormalizeURL(endpoint)
	base = golangsdk.NormalizeURL(base)

	p := new(golangsdk.ProviderClient)
	p.IdentityBase = base
	p.IdentityEndpoint = endpoint
	p.UseTokenLock()

	return p, nil
}

/*
AuthenticatedClient logs in to an OpenStack cloud found at the identity endpoint
specified by the options, acquires a token, and returns a Provider Client
instance that's ready to operate.

If the full path to a versioned identity endpoint was specified  (example:
http://example.com:5000/v3), that path will be used as the endpoint to query.

If a versionless endpoint was specified (example: http://example.com:5000/),
the endpoint will be queried to determine which versions of the identity service
are available, then chooses the most recent or most supported version.

Example:

	ao, err := openstack.AuthOptionsFromEnv()
	provider, err := openstack.AuthenticatedClient(ao)
	client, err := openstack.NewNetworkV2(client, golangsdk.EndpointOpts{
		Region: utils.GetRegion(ao),
	})
*/
func AuthenticatedClient(options golangsdk.AuthOptionsProvider) (*golangsdk.ProviderClient, error) {
	client, err := NewClient(options.GetIdentityEndpoint())
	if err != nil {
		return nil, err
	}

	if err := Authenticate(client, options); err != nil {
		return nil, err
	}
	return client, nil
}

// Authenticate or re-authenticate against the most recent identity service
// supported at the provided endpoint.
func Authenticate(client *golangsdk.ProviderClient, options golangsdk.AuthOptionsProvider) error {
	versions := []*utils.Version{
		{ID: v3, Priority: 30, Suffix: "/v3/"},
	}

	chosen, endpoint, err := utils.ChooseVersion(client, versions)
	if err != nil {
		return err
	}

	authOptions, isTokenAuthOptions := options.(golangsdk.AuthOptions)

	if isTokenAuthOptions {
		switch chosen.ID {
		case v3:
			if authOptions.AgencyDomainName != "" && authOptions.AgencyName != "" {
				return v3authWithAgency(client, endpoint, &authOptions, golangsdk.EndpointOpts{})
			}
			return v3auth(client, endpoint, &authOptions, golangsdk.EndpointOpts{})
		default:
			// The switch statement must be out of date from the versions list.
			return fmt.Errorf("unrecognized identity version: %s", chosen.ID)
		}
	} else {
		akskAuthOptions, isAkSkOptions := options.(golangsdk.AKSKAuthOptions)

		if isAkSkOptions {
			if akskAuthOptions.AgencyDomainName != "" && akskAuthOptions.AgencyName != "" {
				return authWithAgencyByAKSK(client, endpoint, akskAuthOptions, golangsdk.EndpointOpts{})
			}
			return v3AKSKAuth(client, endpoint, akskAuthOptions, golangsdk.EndpointOpts{})

		}
		return fmt.Errorf("unrecognized auth options provider: %s", reflect.TypeOf(options))
	}
}

// AuthenticateV3 explicitly authenticates against the identity v3 service.
func AuthenticateV3(client *golangsdk.ProviderClient, options tokens3.AuthOptionsBuilder, eo golangsdk.EndpointOpts) error {
	return v3auth(client, "", options, eo)
}

type token3Result interface {
	Extract() (*tokens3.Token, error)
	ExtractToken() (*tokens3.Token, error)
	ExtractServiceCatalog() (*tokens3.ServiceCatalog, error)
	ExtractUser() (*tokens3.User, error)
	ExtractRoles() ([]tokens3.Role, error)
	ExtractProject() (*tokens3.Project, error)
}

func v3auth(client *golangsdk.ProviderClient, endpoint string, opts tokens3.AuthOptionsBuilder, eo golangsdk.EndpointOpts) error {
	// Override the generated service endpoint with the one returned by the version endpoint.
	v3Client, err := NewIdentityV3(client, eo)
	if err != nil {
		return err
	}

	if endpoint != "" {
		v3Client.Endpoint = endpoint
	}

	var result token3Result

	if opts.AuthTokenID() != "" { // TODO: Check token validity with Token-By-Token
		v3Client.SetToken(opts.AuthTokenID())
		result = tokens3.Get(v3Client, opts.AuthTokenID())
	} else {
		result = tokens3.Create(v3Client, opts)
	}

	token, err := result.ExtractToken()
	if err != nil {
		return fmt.Errorf("error extracting token: %w", err)
	}

	project, err := result.ExtractProject()
	if err != nil {
		return fmt.Errorf("error extracting project info: %w", err)
	}

	user, err := result.ExtractUser()
	if err != nil {
		return fmt.Errorf("error extracting user info: %w", err)
	}

	serviceCatalog, err := result.ExtractServiceCatalog()
	if err != nil {
		return fmt.Errorf("error extracting service catalog info: %s", err)
	}

	client.TokenID = token.ID
	if project != nil {
		client.ProjectID = project.ID
		client.DomainID = project.Domain.ID
	}
	if user != nil {
		client.UserID = user.ID
		client.DomainID = user.Domain.ID
	}

	if opts.CanReauth() {
		client.ReauthFunc = func() error {
			client.TokenID = ""
			return v3auth(client, endpoint, opts, eo)
		}
	}

	clientRegion := ""
	if aOpts, ok := opts.(*golangsdk.AuthOptions); ok {
		if aOpts.TenantName == "" && project != nil {
			aOpts.TenantName = project.Name
		}
		clientRegion = utils.GetRegion(*aOpts)
	}
	client.RegionID = clientRegion

	client.EndpointLocator = func(opts golangsdk.EndpointOpts) (string, error) {
		// use client region as default one
		if opts.Region == "" && clientRegion != "" {
			opts.Region = clientRegion
		}
		return V3EndpointURL(serviceCatalog, opts)
	}

	return nil
}

func v3authWithAgency(client *golangsdk.ProviderClient, endpoint string, opts *golangsdk.AuthOptions, eo golangsdk.EndpointOpts) error {
	if opts.TokenID == "" {
		err := v3auth(client, endpoint, opts, eo)
		if err != nil {
			return err
		}
	} else {
		client.TokenID = opts.TokenID
	}

	opts1 := golangsdk.AgencyAuthOptions{
		AgencyName:       opts.AgencyName,
		AgencyDomainName: opts.AgencyDomainName,
		DelegatedProject: opts.DelegatedProject,
	}

	return v3auth(client, endpoint, &opts1, eo)
}

func getProjectID(client *golangsdk.ServiceClient, name string) (string, error) {
	opts := projects.ListOpts{
		Name: name,
	}
	allPages, err := projects.List(client, opts).AllPages()
	if err != nil {
		return "", err
	}

	extractProjects, err := projects.ExtractProjects(allPages)

	if err != nil {
		return "", err
	}

	if len(extractProjects) < 1 {
		return "", fmt.Errorf("[DEBUG] cannot find the tenant: %s", name)
	}

	return extractProjects[0].ID, nil
}

func v3AKSKAuth(client *golangsdk.ProviderClient, endpoint string, options golangsdk.AKSKAuthOptions, eo golangsdk.EndpointOpts) error {
	v3Client, err := NewIdentityV3(client, eo)
	if err != nil {
		return err
	}

	// Override the generated service endpoint with the one returned by the version endpoint.
	if endpoint != "" {
		v3Client.Endpoint = endpoint
	}

	// update AKSKAuthOptions of ProviderClient
	// ProviderClient(client) is a reference to the ServiceClient(v3Client)
	defer func() {
		client.AKSKAuthOptions.ProjectId = options.ProjectId
		client.AKSKAuthOptions.DomainID = options.DomainID
	}()

	client.AKSKAuthOptions = options
	client.AKSKAuthOptions.DomainID = ""

	if options.ProjectId == "" && options.ProjectName != "" {
		id, err := getProjectID(v3Client, options.ProjectName)
		if err != nil {
			return err
		}
		options.ProjectId = id
		client.AKSKAuthOptions.ProjectId = options.ProjectId
	}

	if options.DomainID == "" && options.Domain != "" {
		id, err := getDomainID(options.Domain, v3Client)
		if err != nil {
			options.DomainID = ""
		} else {
			options.DomainID = id
		}
	}

	if options.BssDomainID == "" && options.BssDomain != "" {
		id, err := getDomainID(options.BssDomain, v3Client)
		if err != nil {
			options.BssDomainID = ""
		} else {
			options.BssDomainID = id
		}
	}

	client.ProjectID = options.ProjectId
	client.DomainID = options.BssDomainID

	var entries = make([]tokens3.CatalogEntry, 0, 1)
	err = catalog.List(v3Client).EachPage(func(page pagination.Page) (bool, error) {
		catalogList, err := catalog.ExtractServiceCatalog(page)
		if err != nil {
			return false, err
		}

		entries = append(entries, catalogList...)

		return true, nil
	})

	if err != nil {
		return err
	}
	clientRegion := utils.GetRegionFromAKSK(options)
	client.RegionID = clientRegion

	client.EndpointLocator = func(opts golangsdk.EndpointOpts) (string, error) {
		return V3EndpointURL(&tokens3.ServiceCatalog{
			Entries: entries,
		}, opts)
	}
	return nil
}

func authWithAgencyByAKSK(client *golangsdk.ProviderClient, endpoint string, opts golangsdk.AKSKAuthOptions, eo golangsdk.EndpointOpts) error {
	err := v3AKSKAuth(client, endpoint, opts, eo)
	if err != nil {
		return err
	}

	v3Client, err := NewIdentityV3(client, eo)
	if err != nil {
		return err
	}

	if v3Client.AKSKAuthOptions.DomainID == "" {
		return fmt.Errorf("must config domain name")
	}

	opts2 := golangsdk.AgencyAuthOptions{
		AgencyName:       opts.AgencyName,
		AgencyDomainName: opts.AgencyDomainName,
		DelegatedProject: opts.DelegatedProject,
	}
	result := tokens3.Create(v3Client, &opts2)
	token, err := result.ExtractToken()
	if err != nil {
		return err
	}

	project, err := result.ExtractProject()
	if err != nil {
		return fmt.Errorf("error extracting project info: %s", err)
	}

	user, err := result.ExtractUser()
	if err != nil {
		return fmt.Errorf("error extracting user info: %s", err)
	}

	serviceCatalog, err := result.ExtractServiceCatalog()
	if err != nil {
		return err
	}

	client.TokenID = token.ID
	if project != nil {
		client.ProjectID = project.ID
	}
	if user != nil {
		client.UserID = user.ID
	}

	client.ReauthFunc = func() error {
		client.TokenID = ""
		return authWithAgencyByAKSK(client, endpoint, opts, eo)
	}

	client.EndpointLocator = func(opts golangsdk.EndpointOpts) (string, error) {
		return V3EndpointURL(serviceCatalog, opts)
	}

	client.AKSKAuthOptions.AccessKey = ""
	return nil
}

func getDomainID(name string, client *golangsdk.ServiceClient) (string, error) {
	old := client.Endpoint
	defer func() { client.Endpoint = old }()

	client.Endpoint = old + "auth/"

	opts := domains.ListOpts{
		Name: name,
	}
	allPages, err := domains.List(client, &opts).AllPages()
	if err != nil {
		return "", fmt.Errorf("list domains failed, err=%s", err)
	}

	all, err := domains.ExtractDomains(allPages)
	if err != nil {
		return "", fmt.Errorf("extract domains failed, err=%s", err)
	}

	count := len(all)
	switch count {
	case 0:
		err := &golangsdk.ErrResourceNotFound{}
		err.ResourceType = "iam"
		err.Name = name
		return "", err
	case 1:
		return all[0].ID, nil
	default:
		err := &golangsdk.ErrMultipleResourcesFound{}
		err.ResourceType = "iam"
		err.Name = name
		err.Count = count
		return "", err
	}
}

// NewIdentityV3 creates a ServiceClient that may be used to access the v3
// identity service.
func NewIdentityV3(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	endpoint := client.IdentityBase + "v3/"
	clientType := "identity"
	var err error
	if !reflect.DeepEqual(eo, golangsdk.EndpointOpts{}) {
		eo.ApplyDefaults(clientType)
		endpoint, err = client.EndpointLocator(eo)
		if err != nil {
			return nil, err
		}
	}

	// Ensure endpoint still has a suffix of v3.
	// This is because EndpointLocator might have found a versionless
	// endpoint and requests will fail unless targeted at /v3.
	if !strings.HasSuffix(endpoint, "v3/") {
		endpoint = endpoint + "v3/"
	}

	return &golangsdk.ServiceClient{
		ProviderClient: client,
		Endpoint:       endpoint,
		Type:           clientType,
	}, nil
}

func initClientOpts(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts, clientType string) (*golangsdk.ServiceClient, error) {
	sc := new(golangsdk.ServiceClient)
	eo.ApplyDefaults(clientType)
	locator, err := client.EndpointLocator(eo)
	if err != nil {
		return sc, err
	}
	sc.ProviderClient = client
	sc.Endpoint = locator
	sc.Type = clientType
	return sc, nil
}

// initCommonServiceClient is a workaround for services missing from the catalog.
// Firstly, we initialize a service client by "volumev2" type, the endpoint likes https://evs.{region}.{xxx.com}/v2/{project_id}
// then we replace the endpoint with the specified srv and version.
func initCommonServiceClient(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts, srv string, version string) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "volumev2")
	if err != nil {
		return nil, err
	}

	e := strings.Replace(sc.Endpoint, "v2", version, 1)
	sc.Endpoint = strings.Replace(e, "evs", srv, 1)
	sc.ResourceBase = sc.Endpoint
	return sc, err
}

// NewComputeV2 creates a ServiceClient that may be used with the v2 compute
// package.
func NewComputeV2(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "compute")
}

// NewNetworkV2 creates a ServiceClient that may be used with the v2 network
// package.
func NewNetworkV2(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "network")
	if err != nil {
		return nil, err
	}
	sc.ResourceBase = sc.Endpoint + "v2.0/"
	return sc, err
}

// NewBlockStorageV1 creates a ServiceClient that may be used to access the v1
// block storage service.
func NewBlockStorageV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "volume")
}

// NewBlockStorageV2 creates a ServiceClient that may be used to access the v2
// block storage service.
func NewBlockStorageV2(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "volumev2")
}

// NewBlockStorageV3 creates a ServiceClient that may be used to access the v3 block storage service.
func NewBlockStorageV3(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "volumev3")
}

// NewSharedFileSystemV2 creates a ServiceClient that may be used to access the v2 shared file system service.
func NewSharedFileSystemV2(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "sharev2")
}

// NewSharedFileSystemTurboV1 creates a ServiceClient that may be used to access the v2 shared file system turbo service.
func NewSharedFileSystemTurboV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "sfsturbo")
}

// NewOrchestrationV1 creates a ServiceClient that may be used to access the v1
// orchestration service.
func NewOrchestrationV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "orchestration")
}

// NewDNSV2 creates a ServiceClient that may be used to access the v2 DNS
// service.
func NewDNSV2(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "dns")
	if err != nil {
		return nil, err
	}
	sc.ResourceBase = sc.Endpoint + "v2/"
	return sc, err
}

func NewDWSV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "dws")
	if err != nil {
		return nil, err
	}
	sc.ResourceBase = sc.Endpoint + "v1.0/" + client.ProjectID + "/"
	return sc, err
}

func NewIMSV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "image")
	if err != nil {
		return nil, err
	}
	sc.ResourceBase = sc.Endpoint + "v1/"
	return sc, err
}

func NewIMSV2(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "image")
	if err != nil {
		return nil, err
	}
	sc.ResourceBase = sc.Endpoint + "v2/"
	return sc, err
}

// NewOtcV1 creates a ServiceClient that may be used with the v1 network package.
func NewElbV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts, otctype string) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "compute")
	if err != nil {
		return nil, err
	}
	sc.Endpoint = strings.Replace(strings.Replace(sc.Endpoint, "ecs", otctype, 1), "/v2/", "/v1.0/", 1)
	sc.ResourceBase = sc.Endpoint
	sc.Type = otctype
	return sc, err
}

func NewCESClient(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "volumev2")
	if err != nil {
		return nil, err
	}
	e := strings.Replace(sc.Endpoint, "v2", "V1.0", 1)
	sc.Endpoint = strings.Replace(e, "evs", "ces", 1)
	sc.ResourceBase = sc.Endpoint
	return sc, err
}

// NewComputeV1 creates a ServiceClient that may be used with the v1 compute
// package.
func NewComputeV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "ecs")
}

func NewRdsTagV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "network")
	if err != nil {
		return nil, err
	}
	sc.Endpoint = strings.Replace(sc.Endpoint, "vpc", "rds", 1)
	sc.Endpoint = sc.Endpoint + "v1/"
	sc.ResourceBase = sc.Endpoint + client.ProjectID + "/rds/"
	return sc, err
}

// NewAutoScalingV1 creates a ServiceClient that may be used to access the
// auto-scaling service of OpenTelekomCloud public cloud
func NewAutoScalingV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "asv1")
}

// NewAutoScalingV2 creates a ServiceClient that may be used to access the
// auto-scaling service of OpenTelekomCloud public cloud
func NewAutoScalingV2(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "asv1")
	if err != nil {
		return nil, err
	}
	sc.Endpoint = strings.Replace(sc.Endpoint, "v1", "v2", 1)
	return sc, err
}

// NewNetworkV1 creates a ServiceClient that may be used with the v1 network
// package.
func NewNetworkV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "network")
	if err != nil {
		return nil, err
	}
	sc.ResourceBase = sc.Endpoint + "v1/"
	return sc, err
}

// NewVpcEpV1 creates a ServiceClient that may be used with the v1 VPC Endpoint service
func NewVpcEpV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "vpcep")
	if err != nil {
		return nil, err
	}
	return sc, err
}

// NewVpcV3 creates a ServiceClient that may be used with the v3 VPC service
func NewVpcV3(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "vpc")
	if err != nil {
		return nil, err
	}
	sc.Endpoint = strings.Replace(sc.Endpoint, "v1", "v3", 1)
	sc.ResourceBase = sc.Endpoint + "vpc/"
	return sc, err
}

// NewNatV2 creates a ServiceClient that may be used with the v2 nat package.
func NewNatV2(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "nat")
}

// NewMapReduceV1 creates a ServiceClient that may be used with the v1 MapReduce service.
func NewMapReduceV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "mrs")
	if err != nil {
		return nil, err
	}
	sc.ResourceBase = sc.Endpoint + client.ProjectID + "/"
	return sc, err
}

// NewAntiDDoSV1 creates a ServiceClient that may be used with the v1 Anti DDoS Service
// package.
func NewAntiDDoSV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "antiddos")
}

// NewDCaaSV2 creates a ServiceClient that may be used to access the v1 Distributed Message Service.
func NewDCaaSV2(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "dcaas")
}

// NewDMSServiceV1 creates a ServiceClient that may be used to access the v1 Distributed Message Service.
func NewDMSServiceV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "dmsv1")
}

func NewDMSServiceV11(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "dmsv1")
	if err != nil {
		return nil, err
	}
	sc.Endpoint = strings.Replace(sc.Endpoint, "v1.0", "v1", 1)
	return sc, err
}

// NewDMSServiceV2 creates a ServiceClient that may be used to access the v2 Distributed Message Service.
func NewDMSServiceV2(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "dmsv2")
}

// NewDCSServiceV1 creates a ServiceClient that may be used to access the v1 Distributed Cache Service.
func NewDCSServiceV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "network")
	if err != nil {
		return nil, err
	}
	sc.Endpoint = strings.Replace(sc.Endpoint, "vpc", "dcs", 1)
	sc.ResourceBase = sc.Endpoint + "v1.0/" + client.ProjectID + "/"
	return sc, err
}

// NewDDSServiceV3 creates a ServiceClient that may be used to access the Document Database Service.
func NewDDSServiceV3(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "ddsv3")
	return sc, err
}

func NewDISServiceV2(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "dis")
	if err != nil {
		return nil, err
	}
	sc.ResourceBase = sc.Endpoint + "v2/" + client.ProjectID + "/"
	return sc, err
}

// NewDRSServiceV3 creates a ServiceClient that may be used to access the Document Database Service.
func NewDRSServiceV3(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initCommonServiceClient(client, eo, "drs", "v3")
	if err != nil {
		return nil, err
	}
	return sc, nil
}

// NewOBSService creates a ServiceClient that may be used to access the Object Storage Service.
func NewOBSService(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "object")
	return sc, err
}

// NewDeHServiceV1 creates a ServiceClient that may be used to access the v1 Dedicated Hosts service.
func NewDeHServiceV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "deh")
	return sc, err
}

// NewCSBSService creates a ServiceClient that can be used to access the Cloud Server Backup service.
func NewCSBSService(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "data-protect")
	return sc, err
}

// NewCBRService create a ServiceClient that can be used to access the Cloud Backup and Recovery service.
func NewCBRService(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "cbr")
}

// NewCSSService creates a ServiceClient that can be used to access the Cloud Search service.
func NewCSSService(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "css")
}

// NewVBS creates a service client that is used for VBS.
func NewVBS(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "volumev2")
	if err != nil {
		return nil, err
	}
	sc.Endpoint = strings.Replace(sc.Endpoint, "evs", "vbs", 1)
	sc.ResourceBase = sc.Endpoint
	return sc, err
}

func NewVBSServiceV2(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "vbsv2")
}

// NewCTSV1 creates a ServiceClient that can be used to access the Cloud Trace service.
func NewCTSV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "cts")
	return sc, err
}

// NewCTSV2 creates a ServiceClient that can be used to access the Cloud Trace service.
func NewCTSV2(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "cts")
	sc.Endpoint = strings.Replace(sc.Endpoint, "v1", "v2", 1)
	return sc, err
}

// NewCTSV3 creates a ServiceClient that can be used to access the Cloud Trace service.
func NewCTSV3(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "cts")
	sc.Endpoint = strings.Replace(sc.Endpoint, "v1.0", "v3", 1)
	return sc, err
}

// NewELBV1 creates a ServiceClient that may be used to access the ELB service.
func NewELBV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "elbv1")
	return sc, err
}

// NewELBV2 creates a ServiceClient that may be used to access the ELBv2 service.
func NewELBV2(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "elb")
	if suf := "v1.0/"; strings.HasSuffix(sc.Endpoint, suf) {
		sc.Endpoint = strings.TrimSuffix(sc.Endpoint, suf)
	}
	sc.ResourceBase = sc.Endpoint + "v2.0/"
	return sc, err
}

// NewELBV3 creates a ServiceClient that may be used to access the ELBv3 service.
func NewELBV3(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "elbv3")
	if err != nil {
		return nil, err
	}
	sc.ResourceBase = sc.Endpoint + "elb/"
	return sc, nil
}

// NewRDSV1 creates a ServiceClient that may be used to access the RDS service.
func NewRDSV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "rdsv1")
	return sc, err
}

// NewKMSV1 creates a ServiceClient that may be used to access the KMS service.
func NewKMSV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "kmsv1")
}

// NewSMNV2 creates a ServiceClient that may be used to access the SMN service.
func NewSMNV2(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "smnv2")
	if err != nil {
		return nil, err
	}
	sc.ResourceBase = sc.Endpoint + "notifications/"
	return sc, err
}

// NewSMNV2Tags creates a ServiceClient that may be used to access the SMN tags service.
func NewSMNV2Tags(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "smnv2")
	if err != nil {
		return nil, err
	}
	return sc, err
}

// NewCCEv1 creates a ServiceClient that may be used to access the CCE k8s service.
func NewCCEv1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "cce")
}

// NewCCE creates a ServiceClient that may be used to access the CCE service.
func NewCCE(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "ccev2.0")
	if err != nil {
		return nil, err
	}
	sc.ResourceBase = sc.Endpoint + "api/v3/projects/" + client.ProjectID + "/"
	return sc, err
}

// NewWAFV1 creates a ServiceClient that may be used to access the WAF service.
func NewWAFV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "waf")
	if err != nil {
		return nil, err
	}
	sc.ResourceBase = sc.Endpoint + "v1/" + client.ProjectID + "/waf/"
	return sc, err
}

// NewWAFDSwissV1 creates a ServiceClient that may be used to access the premium WAF service.
func NewWAFDSwissV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "waf")
	if err != nil {
		return nil, err
	}
	sc.ResourceBase = sc.Endpoint + "v1/" + client.ProjectID + "/"
	return sc, err
}

// NewWAFDV1 creates a ServiceClient that may be used to access the premium WAF service.
func NewWAFDV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initClientOpts(client, eo, "premium-waf")
	if err != nil {
		return nil, err
	}
	sc.ResourceBase = sc.Endpoint
	return sc, err
}

// NewRDSV3 creates a ServiceClient that may be used to access the RDS service.
func NewRDSV3(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "rdsv3")
}

// NewSDRSV1 creates a ServiceClient that may be used with the v1 SDRS service.
func NewSDRSV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "sdrs")
}

// NewLTSV2 creates a ServiceClient that may be used to access the LTS service.
func NewLTSV2(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initCommonServiceClient(client, eo, "lts", "v2")
	return sc, err
}

// NewSWRV2 creates a ServiceClient that may be used to access the SWR service.
func NewSWRV2(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	serviceClient, err := initClientOpts(client, eo, "smn") // SMN is v2 and has no project ID
	if err != nil {
		return nil, err
	}
	serviceClient.Endpoint = strings.Replace(serviceClient.Endpoint, "smn", "swr-api", 1)
	serviceClient.ResourceBase = serviceClient.Endpoint
	return serviceClient, err
}

// NewTMSV1 creates a ServiceClient that may be used to access the TMS service.
func NewTMSV1(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initClientOpts(client, eo, "tms")
}

func NewGaussDBV3(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initCommonServiceClient(client, eo, "gaussdb", "mysql/v3")
	return sc, err
}

func NewDataArtsV11(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	return initCommonServiceClient(client, eo, "cdm", "v1.1")
}

func NewAPIGW(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initCommonServiceClient(client, eo, "apig", "v2")
	return sc, err
}

func NewFuncGraph(client *golangsdk.ProviderClient, eo golangsdk.EndpointOpts) (*golangsdk.ServiceClient, error) {
	sc, err := initCommonServiceClient(client, eo, "functiongraph", "v2")
	return sc, err
}
