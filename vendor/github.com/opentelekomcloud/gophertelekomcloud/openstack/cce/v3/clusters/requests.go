package clusters

import (
	"reflect"

	"github.com/opentelekomcloud/gophertelekomcloud"
)

var RequestOpts = golangsdk.RequestOpts{
	MoreHeaders: map[string]string{"Content-Type": "application/json"},
}

// ListOpts allows the filtering of list data using given parameters.
type ListOpts struct {
	Name  string `json:"name"`
	ID    string `json:"uuid"`
	Type  string `json:"type"`
	VpcID string `json:"vpc"`
	Phase string `json:"phase"`
}

// List returns collection of clusters.
func List(client *golangsdk.ServiceClient, opts ListOpts) ([]Clusters, error) {
	var r ListResult
	_, r.Err = client.Get(rootURL(client), &r.Body, &golangsdk.RequestOpts{
		OkCodes:     []int{200},
		MoreHeaders: RequestOpts.MoreHeaders, JSONBody: nil,
	})

	allClusters, err := r.ExtractClusters()
	if err != nil {
		return nil, err
	}

	return FilterClusters(allClusters, opts), nil
}

func FilterClusters(clusters []Clusters, opts ListOpts) []Clusters {

	var refinedClusters []Clusters
	var matched bool
	m := map[string]FilterStruct{}

	if opts.Name != "" {
		m["Name"] = FilterStruct{Value: opts.Name, Driller: []string{"Metadata"}}
	}
	if opts.ID != "" {
		m["Id"] = FilterStruct{Value: opts.ID, Driller: []string{"Metadata"}}
	}
	if opts.Type != "" {
		m["Type"] = FilterStruct{Value: opts.Type, Driller: []string{"Spec"}}
	}
	if opts.VpcID != "" {
		m["VpcId"] = FilterStruct{Value: opts.VpcID, Driller: []string{"Spec", "HostNetwork"}}
	}
	if opts.Phase != "" {
		m["Phase"] = FilterStruct{Value: opts.Phase, Driller: []string{"Status"}}
	}

	if len(m) > 0 && len(clusters) > 0 {
		for _, cluster := range clusters {
			matched = true

			for key, value := range m {
				if sVal := GetStructNestedField(&cluster, key, value.Driller); !(sVal == value.Value) {
					matched = false
				}
			}
			if matched {
				refinedClusters = append(refinedClusters, cluster)
			}
		}

	} else {
		refinedClusters = clusters
	}

	return refinedClusters
}

type FilterStruct struct {
	Value   string
	Driller []string
}

func GetStructNestedField(v *Clusters, field string, structDriller []string) string {
	r := reflect.ValueOf(v)
	for _, drillField := range structDriller {
		f := reflect.Indirect(r).FieldByName(drillField).Interface()
		r = reflect.ValueOf(f)
	}
	f1 := reflect.Indirect(r).FieldByName(field)
	return f1.String()
}

// CreateOptsBuilder allows extensions to add additional parameters to the
// Create request.
type CreateOptsBuilder interface {
	ToClusterCreateMap() (map[string]interface{}, error)
}

// CreateOpts contains all the values needed to create a new cluster
type CreateOpts struct {
	// API type, fixed value Cluster
	Kind string `json:"kind" required:"true"`
	// API version, fixed value v3
	ApiVersion string `json:"apiversion" required:"true"`
	// Metadata required to create a cluster
	Metadata CreateMetaData `json:"metadata" required:"true"`
	// specifications to create a cluster
	Spec Spec `json:"spec" required:"true"`
}

// Metadata required to create a cluster
type CreateMetaData struct {
	// Cluster unique name
	Name string `json:"name" required:"true"`
	// Cluster tag, key/value pair format
	Labels map[string]string `json:"labels,omitempty"`
	// Cluster annotation, key/value pair format
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ToClusterCreateMap builds a create request body from CreateOpts.
func (opts CreateOpts) ToClusterCreateMap() (map[string]interface{}, error) {
	return golangsdk.BuildRequestBody(opts, "")
}

type ExpirationOptsBuilder interface {
	ToExpirationGetMap() (map[string]interface{}, error)
}

type ExpirationOpts struct {
	Duration int `json:"duration" required:"true"`
}

func (opts ExpirationOpts) ToExpirationGetMap() (map[string]interface{}, error) {
	return golangsdk.BuildRequestBody(opts, "")
}

// Create accepts a CreateOpts struct and uses the values to create a new
// logical cluster.
func Create(c *golangsdk.ServiceClient, opts CreateOptsBuilder) (r CreateResult) {
	b, err := opts.ToClusterCreateMap()
	if err != nil {
		r.Err = err
		return
	}
	reqOpt := &golangsdk.RequestOpts{OkCodes: []int{201}}
	_, r.Err = c.Post(rootURL(c), b, &r.Body, reqOpt)
	return
}

// Get retrieves a particular cluster based on its unique ID.
func Get(c *golangsdk.ServiceClient, id string) (r GetResult) {
	_, r.Err = c.Get(resourceURL(c, id), &r.Body, &golangsdk.RequestOpts{
		OkCodes:     []int{200},
		MoreHeaders: RequestOpts.MoreHeaders, JSONBody: nil,
	})
	return
}

// GetCert retrieves a particular cluster certificate based on its unique ID.
func GetCert(c *golangsdk.ServiceClient, id string) (r GetCertResult) {
	_, r.Err = c.Get(certificateURL(c, id), &r.Body, &golangsdk.RequestOpts{
		OkCodes:     []int{200},
		MoreHeaders: RequestOpts.MoreHeaders,
	})
	return
}

// GetCertWithExpiration retrieves a particular cluster certificate based on its unique ID.
func GetCertWithExpiration(c *golangsdk.ServiceClient, id string, opts ExpirationOptsBuilder) (r GetCertResult) {
	b, err := opts.ToExpirationGetMap()
	if err != nil {
		r.Err = err
		return
	}

	_, r.Err = c.Post(certificateURL(c, id), b, &r.Body, &golangsdk.RequestOpts{
		OkCodes:     []int{200},
		MoreHeaders: RequestOpts.MoreHeaders,
	})
	return
}

// UpdateOpts contains all the values needed to update a new cluster
type UpdateOpts struct {
	Spec UpdateSpec `json:"spec" required:"true"`
}

type UpdateSpec struct {
	// Cluster description
	Description string `json:"description,omitempty"`
}

// UpdateOptsBuilder allows extensions to add additional parameters to the
// Update request.
type UpdateOptsBuilder interface {
	ToClusterUpdateMap() (map[string]interface{}, error)
}

// ToClusterUpdateMap builds an update body based on UpdateOpts.
func (opts UpdateOpts) ToClusterUpdateMap() (map[string]interface{}, error) {
	return golangsdk.BuildRequestBody(opts, "")
}

// Update allows clusters to update description.
func Update(c *golangsdk.ServiceClient, id string, opts UpdateOptsBuilder) (r UpdateResult) {
	b, err := opts.ToClusterUpdateMap()
	if err != nil {
		r.Err = err
		return
	}
	_, r.Err = c.Put(resourceURL(c, id), b, &r.Body, &golangsdk.RequestOpts{
		OkCodes: []int{200},
	})
	return
}

// Delete will permanently delete a particular cluster based on its unique ID.
func Delete(c *golangsdk.ServiceClient, id string) (r DeleteResult) {
	_, r.Err = c.Delete(resourceURL(c, id), &golangsdk.RequestOpts{
		OkCodes:     []int{200},
		MoreHeaders: RequestOpts.MoreHeaders, JSONBody: nil,
	})
	return
}

type DeleteOpts struct {
	ErrorStatus string `q:"errorStatus"`
	DeleteEfs   string `q:"delete_efs"`
	DeleteENI   string `q:"delete_eni"`
	DeleteEvs   string `q:"delete_evs"`
	DeleteNet   string `q:"delete_net"`
	DeleteObs   string `q:"delete_obs"`
	DeleteSfs   string `q:"delete_sfs"`
}

func DeleteWithOpts(c *golangsdk.ServiceClient, id string, opts DeleteOpts) error {
	url := resourceURL(c, id)
	q, err := golangsdk.BuildQueryString(&opts)
	if err != nil {
		return err
	}

	_, err = c.Delete(url+q.String(), &golangsdk.RequestOpts{
		OkCodes:     []int{200},
		MoreHeaders: RequestOpts.MoreHeaders, JSONBody: nil,
	})
	return err
}

type UpdateIpOpts struct {
	Action    string `json:"action" required:"true"`
	Spec      IpSpec `json:"spec,omitempty"`
	ElasticIp string `json:"elasticIp"`
}

type IpSpec struct {
	ID string `json:"id" required:"true"`
}

type UpdateIpOptsBuilder interface {
	ToMasterIpUpdateMap() (map[string]interface{}, error)
}

func (opts UpdateIpOpts) ToMasterIpUpdateMap() (map[string]interface{}, error) {
	return golangsdk.BuildRequestBody(opts, "spec")
}

// Update the access information of a specified cluster.
func UpdateMasterIp(c *golangsdk.ServiceClient, id string, opts UpdateIpOptsBuilder) (r UpdateIpResult) {
	b, err := opts.ToMasterIpUpdateMap()
	if err != nil {
		r.Err = err
		return
	}
	_, r.Err = c.Put(masterIpURL(c, id), b, &r.Body, &golangsdk.RequestOpts{
		OkCodes: []int{200},
	})
	return
}
