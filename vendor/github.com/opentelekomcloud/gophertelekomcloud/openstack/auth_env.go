package openstack

import (
	"github.com/opentelekomcloud/gophertelekomcloud"
)

/*
AuthOptionsFromEnv fills out an identity.AuthOptions structure with the
settings found on the various OpenStack OS_* environment variables.

The following variables provide sources of truth: OS_AUTH_URL, OS_USERNAME,
OS_PASSWORD, OS_TENANT_ID, and OS_TENANT_NAME.

Of these, OS_USERNAME, OS_PASSWORD, and OS_AUTH_URL must have settings,
or an error will result.  OS_TENANT_ID, OS_TENANT_NAME, OS_PROJECT_ID, and
OS_PROJECT_NAME are optional.

OS_TENANT_ID and OS_TENANT_NAME are mutually exclusive to OS_PROJECT_ID and
OS_PROJECT_NAME. If OS_PROJECT_ID and OS_PROJECT_NAME are set, they will
still be referred as "tenant" in Gophercloud.

To use this function, first set the OS_* environment variables (for example,
by sourcing an `openrc` file), then:

	opts, err := openstack.AuthOptionsFromEnv()
	provider, err := openstack.AuthenticatedClient(opts)
*/
func AuthOptionsFromEnv(envs ...*Env) (golangsdk.AuthOptions, error) {
	e := NewEnv(defaultPrefix)
	if len(envs) > 0 {
		e = envs[0]
	}

	ao := golangsdk.AuthOptions{
		IdentityEndpoint: e.GetEnv("AUTH_URL"),
		Username:         e.GetEnv("USERNAME"),
		UserID:           e.GetEnv("USERID", "USER_ID"),
		Password:         e.GetEnv("PASSWORD"),
		DomainID:         e.GetEnv("DOMAIN_ID", "USER_DOMAIN_ID", "PROJECT_DOMAIN_ID"),
		DomainName:       e.GetEnv("DOMAIN_NAME", "USER_DOMAIN_NAME", "PROJECT_DOMAIN_NAME"),
		TenantID:         e.GetEnv("PROJECT_ID", "TENANT_ID"),
		TenantName:       e.GetEnv("PROJECT_NAME", "TENANT_NAME"),
		TokenID:          e.GetEnv("TOKEN", "TOKEN_ID"),
		AgencyName:       e.GetEnv("AGENCY_NAME", "TARGET_AGENCY_NAME"),
		AgencyDomainName: e.GetEnv("AGENCY_DOMAIN_NAME", "TARGET_DOMAIN_NAME"),
		DelegatedProject: e.GetEnv("DELEGATED_PROJECT", "TARGET_DOMAIN_NAME"),
		Passcode:         e.GetEnv("PASSCODE"),
	}
	return ao, nil
}
