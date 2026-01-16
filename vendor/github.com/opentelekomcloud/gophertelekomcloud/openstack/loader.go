package openstack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/utils"
)

const (
	defaultEnvVarKey = "envvars"
	defaultPrefix    = "OS_"

	DefaultProfileName = "otc"
	regionPlaceHolder  = "{region_name}"
)

var (
	yamlSuffixes = []string{".yaml", ".yml"}
	jsonSuffixes = []string{".json"}

	configFiles = fileList("clouds")
	secureFiles = fileList("secure")
	vendorFiles = fileList("clouds-public")

	OTCVendorConfig = &VendorConfig{
		Clouds: map[string]Cloud{
			DefaultProfileName: {
				AuthInfo: AuthInfo{
					AuthURL: fmt.Sprintf("https://iam.%s.otc.t-systems.com/v3", regionPlaceHolder),
				},
				Regions:            []string{"eu-de", "eu-nl"},
				EndpointType:       "public",
				IdentityAPIVersion: "3",
			},
		},
	}
)

func configSearchPath() []string {
	home, _ := os.UserHomeDir()
	cwd, _ := os.Getwd()
	userConfigDir, _ := filepath.Abs(filepath.Join(home, ".config/openstack"))
	unixConfigDir, _ := filepath.Abs("/etc/openstack")
	return []string{
		cwd,
		userConfigDir,
		unixConfigDir,
	}
}

func fileList(name string) []string {
	paths := configSearchPath()
	var suffixes []string
	suffixes = append(suffixes, yamlSuffixes...)
	suffixes = append(suffixes, jsonSuffixes...)
	size := len(suffixes) * len(paths)
	files := make([]string, size)
	i := 0
	for _, path := range paths {
		for _, suffix := range suffixes {
			files[i] = filepath.Join(path, name+suffix)
			i++
		}
	}
	return files
}

type Env struct {
	// prefix of the invironment, `OS_` in most cases
	prefix string
	// cloud containins all information about used cloud
	cloud *Cloud
	// unstable make Env ignore lazy cloud loading and
	// refresh it every time it's requested
	unstable bool
}

// NewEnv create new <prefixed> Env loader, lazy by default
func NewEnv(prefix string, lazy ...bool) *Env {
	if prefix != "" && !strings.HasSuffix(prefix, "_") {
		prefix += "_"
	}
	unstable := false
	if len(lazy) > 0 {
		unstable = !lazy[0]
	}
	return &Env{prefix: prefix, unstable: unstable}
}

func (e *Env) Prefix() string {
	return e.prefix
}

func (e *Env) cloudFromEnv() *Cloud {
	authOpts, _ := AuthOptionsFromEnv(e)
	verify := true
	if v := e.GetEnv("INSECURE"); v != "" {
		verify = v != "1" && v != "true"
	}
	aws := NewEnv("AWS_")
	access := aws.GetEnv("ACCESS_KEY_ID")
	if access == "" {
		access = e.GetEnv("ACCESS_KEY", "ACCESS_KEY_ID", "AK")
	}
	secret := aws.GetEnv("ACCESS_SECRET_KEY")
	if secret == "" {
		secret = e.GetEnv("SECRET_KEY", "ACCESS_KEY_SECRET", "SK")
	}
	security := aws.GetEnv("SECURITY_TOKEN")
	if security == "" {
		security = e.GetEnv("SECURITY_TOKEN", "AKSK_SECURITY_TOKEN", "ST")
	}
	region := e.GetEnv("REGION_NAME", "REGION_ID")
	if region == "" {
		region = utils.GetRegion(authOpts)
	}

	cloud := &Cloud{
		Cloud:   e.GetEnv("CLOUD"),
		Profile: e.GetEnv("PROFILE"),
		AuthInfo: AuthInfo{
			AuthURL:           authOpts.IdentityEndpoint,
			Token:             authOpts.TokenID,
			Username:          authOpts.Username,
			UserID:            authOpts.UserID,
			Password:          authOpts.Password,
			Passcode:          authOpts.Passcode,
			ProjectName:       authOpts.TenantName,
			ProjectID:         authOpts.TenantID,
			UserDomainName:    e.GetEnv("USER_DOMAIN_NAME"),
			UserDomainID:      e.GetEnv("USER_DOMAIN_ID"),
			ProjectDomainName: e.GetEnv("PROJECT_DOMAIN_NAME"),
			ProjectDomainID:   e.GetEnv("PROJECT_DOMAIN_ID"),
			DomainName:        authOpts.DomainName,
			DomainID:          authOpts.DomainID,
			DefaultDomain:     e.GetEnv("DEFAULT_DOMAIN"),
			AccessKey:         access,
			SecretKey:         secret,
			SecurityToken:     security,
			AgencyName:        authOpts.AgencyName,
			AgencyDomainName:  authOpts.AgencyDomainName,
			DelegatedProject:  authOpts.DelegatedProject,
		},
		AuthType:           AuthType(e.GetEnv("AUTH_TYPE")),
		RegionName:         region,
		EndpointType:       e.GetEnv("ENDPOINT_TYPE"),
		Interface:          e.GetEnv("INTERFACE"),
		IdentityAPIVersion: e.GetEnv("IDENTITY_API_VERSION"),
		VolumeAPIVersion:   e.GetEnv("VOLUME_API_VERSION"),
		Verify:             &verify,
		CACertFile:         e.GetEnv("CA_CERT", "CA_CERT_FILE"),
		ClientCertFile:     e.GetEnv("CLIENT_CERT", "CLIENT_CERT_FILE"),
		ClientKeyFile:      e.GetEnv("CLIENT_KEY", "CLIENT_KEY_FILE"),
	}
	return cloud
}

// GetEnv returns first non-empty value of given environment variables
func (e *Env) GetEnv(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(e.prefix + key); value != "" {
			return value
		}
	}
	return ""
}

// VendorConfig represents a collection of PublicCloud entries in clouds-public.yaml file.
// The format of the clouds-public.yml is documented at
// https://docs.openstack.org/python-openstackclient/latest/configuration/
type VendorConfig struct {
	Clouds map[string]Cloud `yaml:"public-clouds" json:"public-clouds"`
}

// Config represents a collection of Cloud entries in a clouds.yaml file.
// The format of clouds.yaml is documented at
// https://docs.openstack.org/os-client-config/latest/user/configuration.html.
type Config struct {
	DefaultCloud string           `yaml:"-" json:"-"`
	Clouds       map[string]Cloud `yaml:"clouds" json:"clouds"`
}

func NewConfig() *Config {
	return &Config{
		Clouds: map[string]Cloud{},
	}
}

// AuthType represents a valid method of authentication: `password`, `token`, `aksk` or `agency`
type AuthType string

// AuthInfo represents the auth section of a cloud entry
type AuthInfo struct {
	// AuthURL is the keystone/identity endpoint URL.
	AuthURL string `yaml:"auth_url,omitempty" json:"auth_url,omitempty"`

	// Token is a pre-generated authentication token.
	Token string `yaml:"token,omitempty" json:"token,omitempty"`

	// Username is the username of the user.
	Username string `yaml:"username,omitempty" json:"username,omitempty"`

	// UserID is the unique ID of a user.
	UserID string `yaml:"user_id,omitempty" json:"user_id,omitempty"`

	// Password is the password of the user.
	Password string `yaml:"password,omitempty" json:"password,omitempty"`

	// Passcode for MFA.
	Passcode string `yaml:"-" json:"-"`

	// ProjectName is the common/human-readable name of a project.
	// Users can be scoped to a project.
	// ProjectName on its own is not enough to ensure a unique scope. It must
	// also be combined with either a ProjectDomainName or ProjectDomainID.
	// ProjectName cannot be combined with ProjectID in a scope.
	ProjectName string `yaml:"project_name,omitempty" json:"project_name,omitempty"`

	// ProjectID is the unique ID of a project.
	// It can be used to scope a user to a specific project.
	ProjectID string `yaml:"project_id,omitempty" json:"project_id,omitempty"`

	// UserDomainName is the name of the domain where a user resides.
	// It is used to identify the source domain of a user.
	UserDomainName string `yaml:"user_domain_name,omitempty" json:"user_domain_name,omitempty"`

	// UserDomainID is the unique ID of the domain where a user resides.
	// It is used to identify the source domain of a user.
	UserDomainID string `yaml:"user_domain_id,omitempty" json:"user_domain_id,omitempty"`

	// ProjectDomainName is the name of the domain where a project resides.
	// It is used to identify the source domain of a project.
	// ProjectDomainName can be used in addition to a ProjectName when scoping
	// a user to a specific project.
	ProjectDomainName string `yaml:"project_domain_name,omitempty" json:"project_domain_name,omitempty"`

	// ProjectDomainID is the name of the domain where a project resides.
	// It is used to identify the source domain of a project.
	// ProjectDomainID can be used in addition to a ProjectName when scoping
	// a user to a specific project.
	ProjectDomainID string `yaml:"project_domain_id,omitempty" json:"project_domain_id,omitempty"`

	// DomainName is the name of a domain which can be used to identify the
	// source domain of either a user or a project.
	// If UserDomainName and ProjectDomainName are not specified, then DomainName
	// is used as a default choice.
	// It can also be used be used to specify a domain-only scope.
	DomainName string `yaml:"domain_name,omitempty" json:"domain_name,omitempty"`

	// DomainID is the unique ID of a domain which can be used to identify the
	// source domain of either a user or a project.
	// If UserDomainID and ProjectDomainID are not specified, then DomainID is
	// used as a default choice.
	// It can also be used be used to specify a domain-only scope.
	DomainID string `yaml:"domain_id,omitempty" json:"domain_id,omitempty"`

	// DefaultDomain is the domain ID to fall back on if no other domain has
	// been specified and a domain is required for scope.
	DefaultDomain string `yaml:"default_domain,omitempty" json:"default_domain,omitempty"`

	// AK/SK auth means
	AccessKey     string `yaml:"ak,omitempty" json:"ak,omitempty"`
	SecretKey     string `yaml:"sk,omitempty" json:"sk,omitempty"`
	SecurityToken string `yaml:"security_token,omitempty" json:"security_token,omitempty"`

	// OTC Agency config
	AgencyName string `yaml:"target_agency_name,omitempty" json:"agency_name,omitempty"`
	// AgencyDomainName is the name of domain who created the agency
	AgencyDomainName string `yaml:"target_domain_id,omitempty" json:"target_domain_id,omitempty"`
	// DelegatedProject is the name of delegated project
	DelegatedProject string `yaml:"target_project_name,omitempty" json:"target_project_name,omitempty"`
}

// Cloud represents an entry in a clouds.yaml/public-clouds.yaml/secure.yaml file.
type Cloud struct {
	Cloud      string   `yaml:"cloud,omitempty" json:"cloud,omitempty"`
	Profile    string   `yaml:"profile,omitempty" json:"profile,omitempty"`
	AuthType   AuthType `yaml:"auth_type,omitempty" json:"auth_type,omitempty"`
	AuthInfo   AuthInfo `yaml:"auth,omitempty" json:"auth,omitempty"`
	RegionName string   `yaml:"region_name,omitempty" json:"region_name,omitempty"`
	Regions    []string `yaml:"regions,omitempty" json:"regions,omitempty"`

	// EndpointType and Interface both specify whether to use the public, internal,
	// or admin interface of a service. They should be considered synonymous, but
	// EndpointType will take precedence when both are specified.
	EndpointType string `yaml:"endpoint_type,omitempty" json:"endpoint_type,omitempty"`
	Interface    string `yaml:"interface,omitempty" json:"interface,omitempty"`

	// API Version overrides.
	IdentityAPIVersion string `yaml:"identity_api_version,omitempty" json:"identity_api_version,omitempty"`
	VolumeAPIVersion   string `yaml:"volume_api_version,omitempty" json:"volume_api_version,omitempty"`

	// Verify whether or not SSL API requests should be verified.
	Verify *bool `yaml:"verify,omitempty" json:"verify,omitempty"`

	// CACertFile a path to a CA Cert bundle that can be used as part of
	// verifying SSL API requests.
	CACertFile string `yaml:"cacert,omitempty" json:"cacert,omitempty"`

	// ClientCertFile a path to a client certificate to use as part of the SSL
	// transaction.
	ClientCertFile string `yaml:"cert,omitempty" json:"cert,omitempty"`

	// ClientKeyFile a path to a client key to use as part of the SSL
	// transaction.
	ClientKeyFile string `yaml:"key,omitempty" json:"key,omitempty"`
}

func (c *Cloud) computeRegion() {
	if c.RegionName != "" {
		return
	}
	name := c.AuthInfo.ProjectName
	if name == "" {
		name = c.AuthInfo.DelegatedProject
	}
	c.RegionName = strings.Split(name, "_")[0]
}

func (c *Cloud) computeAuthURL() error {
	// Auth URL depends on provided region
	if url := c.AuthInfo.AuthURL; strings.Contains(url, regionPlaceHolder) {
		if c.RegionName == "" {
			return fmt.Errorf("region placeholder found in `AuthURL` (%s), but no region provided", url)
		}
		c.AuthInfo.AuthURL = strings.ReplaceAll(url, regionPlaceHolder, c.RegionName)
	}
	return nil
}

func loadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func loadCloudFile(path string) (*Config, error) {
	data, err := loadFile(path)
	if err != nil {
		return nil, err
	}
	clouds := NewConfig()
	if err := yaml.Unmarshal(data, clouds); err != nil {
		return nil, err
	}
	return clouds, err
}

func loadVendorFile(path string) (*VendorConfig, error) {
	data, err := loadFile(path)
	if err != nil {
		return nil, err
	}
	clouds := new(VendorConfig)
	if err := yaml.Unmarshal(data, clouds); err != nil {
		return nil, err
	}
	return clouds, err
}

func mergeWithVendor(config *Config, vendor *VendorConfig) (*Config, error) {
	for k, cloud := range config.Clouds {
		profile := cloud.Profile
		if profile == "" {
			profile = cloud.Cloud
		}
		if profile == "" {
			continue
		}
		if v, ok := vendor.Clouds[profile]; ok {
			merged, err := mergeClouds(&cloud, &v)
			if err != nil {
				log.Printf("error during merge with vendor file: %s", err)
				return config, err
			}
			config.Clouds[k] = *merged
		}
	}
	return config, nil
}

func mergeCloudConfigs(config, fallback *Config) (*Config, error) {
	resultClouds := &Config{
		Clouds: map[string]Cloud{},
	}
	for profile, cfg := range config.Clouds {
		if fallback, ok := fallback.Clouds[profile]; ok {
			cld, err := mergeClouds(cfg, fallback)
			if err != nil {
				return nil, err
			}
			resultClouds.Clouds[profile] = *cld
		} else {
			resultClouds.Clouds[profile] = cfg
		}
	}
	return resultClouds, nil
}

func selectExisting(files []string) string {
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			return file
		}
	}
	return ""
}

// mergeClouds merges two Config recursively (the AuthInfo also gets merged).
// In case both Config define a value, the value in the 'cloud' cloud takes precedence
func mergeClouds(cloud, fallback interface{}) (*Cloud, error) {
	overrideJson, err := json.Marshal(fallback)
	if err != nil {
		return nil, err
	}
	cloudJson, err := json.Marshal(cloud)
	if err != nil {
		return nil, err
	}
	var fallbackInterface interface{}
	err = json.Unmarshal(overrideJson, &fallbackInterface)
	if err != nil {
		return nil, err
	}
	var cloudInterface interface{}
	err = json.Unmarshal(cloudJson, &cloudInterface)
	if err != nil {
		return nil, err
	}
	var mergedCloud Cloud
	mergedInterface := utils.MergeInterfaces(cloudInterface, fallbackInterface)
	mergedJson, err := json.Marshal(mergedInterface)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(mergedJson, &mergedCloud); err != nil {
		return nil, err
	}
	return &mergedCloud, nil
}

func mergeWithSecure(cloudConfig *Config, securePath string) *Config {
	s, err := loadCloudFile(securePath)
	if err != nil {
		log.Printf("Failed to load %s as secure config", securePath)
		return cloudConfig
	}
	cc, err := mergeCloudConfigs(cloudConfig, s)
	if err != nil {
		log.Printf("Failed to merge %s into cloud config", securePath)
		return cloudConfig
	}
	return cc
}

func mergeWithVendors(cloudConfig *Config, vendorPath string) *Config {
	v, err := loadVendorFile(vendorPath)
	if err != nil {
		log.Printf("Failed to load %s as vendor config", vendorPath)
		return cloudConfig
	}
	cc, err := mergeWithVendor(cloudConfig, v)
	if err != nil {
		log.Printf("Failed to merge %s into vendor config", vendorPath)
		return cloudConfig
	}
	return cc
}

// Cloud get cloud merged from configuration and env variables
// if `cloudName` is not empty, explicit cloud name will be used instead
// defined in `OS_CLOUD` environment variable
func (e *Env) Cloud(name ...string) (*Cloud, error) {
	cloudName := ""
	if len(name) > 0 {
		cloudName = name[0]
		e.unstable = true // previously loaded cloud can be different
	}
	if e.cloud == nil || e.unstable {
		config, err := e.loadOpenstackConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load clouds configuration: %s", err)
		}
		if cloudName == "" {
			cloudName = config.DefaultCloud
		}
		cloud, err := mergeClouds(
			config.Clouds[cloudName],
			e.cloudFromEnv(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to merge cloud %s with Env vars: %s", config.DefaultCloud, err)
		}
		cloud.Cloud = cloudName // override value read from environment
		cloud.computeRegion()
		if err := cloud.computeAuthURL(); err != nil {
			return nil, err
		}
		e.cloud = cloud
	}
	return e.cloud, nil

}

// LoadCloudConfig utilize all existing cloud configurations to create cloud configuration:
// env variables, clouds.yaml, secure.yaml, clouds-public.yaml
func (e *Env) loadOpenstackConfig() (*Config, error) {
	var (
		configs = make([]string, len(configFiles))
		secure  = make([]string, len(secureFiles))
		vendors = make([]string, len(vendorFiles))
	)
	copy(configs, configFiles)
	copy(secure, secureFiles)
	copy(vendors, vendorFiles)

	// find config files
	if c := e.GetEnv("CLIENT_CONFIG_FILE"); c != "" {
		configs = utils.PrependString(c, configs)
	}
	configPath := selectExisting(configs)

	if s := e.GetEnv("CLIENT_SECURE_FILE"); s != "" {
		secure = utils.PrependString(s, secure)
	}
	securePath := selectExisting(secure)

	if v := e.GetEnv("CLIENT_VENDOR_FILE"); v != "" {
		vendors = utils.PrependString(v, vendors)
	}
	vendorPath := selectExisting(vendors)

	cloudConfig := NewConfig()

	// load clouds.yaml
	if configPath != "" {
		c, err := loadCloudFile(configPath)
		if err != nil {
			log.Printf("Failed to load %s as cloud config", securePath)
		}
		if c.Clouds != nil {
			cloudConfig = c
		}
	}

	// merge with secure.yaml
	if securePath != "" {
		cloudConfig = mergeWithSecure(cloudConfig, securePath)
	}

	// append cloud from envvars
	envVarKey := e.GetEnv("CLOUD_NAME")
	if envVarKey == "" {
		envVarKey = defaultEnvVarKey
	}
	if _, ok := cloudConfig.Clouds[envVarKey]; ok {
		return nil, fmt.Errorf("%sCLOUD_NAME=`%s` duplicates cloud defined in file", e.prefix, envVarKey)
	}
	cloudConfig.Clouds[envVarKey] = *NewEnv(envVarKey).cloudFromEnv()

	cloudName := e.GetEnv("CLOUD")
	if cloudName == "" && len(cloudConfig.Clouds) == 1 {
		for k := range cloudConfig.Clouds {
			cloudName = k
		}
	}
	cloudConfig.DefaultCloud = cloudName

	// merge with clouds-public.yaml
	var err error
	if vendorPath != "" {
		cloudConfig = mergeWithVendors(cloudConfig, vendorPath)
	} else {
		cloudConfig, err = mergeWithVendor(cloudConfig, OTCVendorConfig)
	}
	return cloudConfig, err
}

func getAuthType(val AuthType) AuthType {
	explicitTypes := []string{"token", "password", "aksk"}
	for _, opt := range explicitTypes {
		if strings.Contains(string(val), opt) {
			return AuthType(opt)
		}
	}
	return val
}

// AuthOptionsFromInfo builds auth options from auth info and type. Returns either AuthOptions or AKSKAuthOptions
func AuthOptionsFromInfo(authInfo *AuthInfo, authType AuthType) (golangsdk.AuthOptionsProvider, error) {
	// project scope
	if authInfo.ProjectID != "" || authInfo.ProjectName != "" {
		if authInfo.ProjectDomainName != "" {
			authInfo.DomainName = authInfo.ProjectDomainName
		}
		if authInfo.ProjectDomainID != "" {
			authInfo.ProjectID = authInfo.ProjectDomainID
		}
	}
	// user scope
	if authInfo.Username != "" || authInfo.UserID != "" {
		if authInfo.UserDomainName != "" {
			authInfo.DomainName = authInfo.UserDomainName
		}
		if authInfo.UserDomainID != "" {
			authInfo.ProjectID = authInfo.UserDomainID
		}
	}

	ao := golangsdk.AuthOptions{
		IdentityEndpoint: authInfo.AuthURL,
		TokenID:          authInfo.Token,
		Username:         authInfo.Username,
		UserID:           authInfo.UserID,
		Password:         authInfo.Password,
		DomainID:         authInfo.DomainID,
		DomainName:       authInfo.DomainName,
		TenantID:         authInfo.ProjectID,
		TenantName:       authInfo.ProjectName,
		Passcode:         authInfo.Passcode,
	}

	explicitAuthType := getAuthType(authType)

	// If an auth_type of "token" was specified, then make sure
	// Gophercloud properly authenticates with a token. This involves
	// unsetting a few other auth options. The reason this is done
	// here is to wait until all auth settings (both in clouds.yaml
	// and via environment variables) are set and then unset them.
	if explicitAuthType == "token" || explicitAuthType == "aksk" {
		ao.Username = ""
		ao.Password = ""
		ao.UserID = ""
		ao.DomainID = ""
		ao.DomainName = ""
	}

	// Check for absolute minimum requirements.
	if ao.IdentityEndpoint == "" {
		err := golangsdk.ErrMissingInput{Argument: "auth_url"}
		return nil, err
	}
	if explicitAuthType == "aksk" || (explicitAuthType == "" && authInfo.AccessKey != "") {
		return golangsdk.AKSKAuthOptions{
			IdentityEndpoint: ao.IdentityEndpoint,
			ProjectId:        ao.TenantID,
			ProjectName:      ao.TenantName,
			Domain:           ao.DomainName,
			DomainID:         ao.DomainID,
			AccessKey:        authInfo.AccessKey,
			SecretKey:        authInfo.SecretKey,
			AgencyName:       ao.AgencyName,
			AgencyDomainName: ao.AgencyDomainName,
			DelegatedProject: ao.DelegatedProject,
		}, nil
	}
	return ao, nil
}

// AuthenticatedClient create new client based on used Env prefix
// this uses LoadOpenstackConfig inside
func (e *Env) AuthenticatedClient(cloudName ...string) (*golangsdk.ProviderClient, error) {
	cloud, err := e.Cloud(cloudName...)
	if err != nil {
		return nil, err
	}
	return AuthenticatedClientFromCloud(cloud)
}

// AuthenticatedClientFromCloud create new authenticated client for given cloud config
func AuthenticatedClientFromCloud(cloud *Cloud) (*golangsdk.ProviderClient, error) {
	opts, err := AuthOptionsFromInfo(&cloud.AuthInfo, cloud.AuthType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert AuthInfo to AuthOptsBuilder with Env vars: %s", err)
	}
	client, err := AuthenticatedClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate client: %s", err)
	}
	return client, nil
}
