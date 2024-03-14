// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

// Package k8s provides methods and message types of the k8s v1 API.
package k8s

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/scaleway/scaleway-sdk-go/internal/errors"
	"github.com/scaleway/scaleway-sdk-go/internal/marshaler"
	"github.com/scaleway/scaleway-sdk-go/internal/parameter"
	"github.com/scaleway/scaleway-sdk-go/namegenerator"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// always import dependencies
var (
	_ fmt.Stringer
	_ json.Unmarshaler
	_ url.URL
	_ net.IP
	_ http.Header
	_ bytes.Reader
	_ time.Time
	_ = strings.Join

	_ scw.ScalewayRequest
	_ marshaler.Duration
	_ scw.File
	_ = parameter.AddToQuery
	_ = namegenerator.GetRandomName
)

// API: kubernetes API.
type API struct {
	client *scw.Client
}

// NewAPI returns a API object from a Scaleway client.
func NewAPI(client *scw.Client) *API {
	return &API{
		client: client,
	}
}

type AutoscalerEstimator string

const (
	AutoscalerEstimatorUnknownEstimator = AutoscalerEstimator("unknown_estimator")
	AutoscalerEstimatorBinpacking       = AutoscalerEstimator("binpacking")
)

func (enum AutoscalerEstimator) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown_estimator"
	}
	return string(enum)
}

func (enum AutoscalerEstimator) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *AutoscalerEstimator) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = AutoscalerEstimator(AutoscalerEstimator(tmp).String())
	return nil
}

type AutoscalerExpander string

const (
	AutoscalerExpanderUnknownExpander = AutoscalerExpander("unknown_expander")
	AutoscalerExpanderRandom          = AutoscalerExpander("random")
	AutoscalerExpanderMostPods        = AutoscalerExpander("most_pods")
	AutoscalerExpanderLeastWaste      = AutoscalerExpander("least_waste")
	AutoscalerExpanderPriority        = AutoscalerExpander("priority")
	AutoscalerExpanderPrice           = AutoscalerExpander("price")
)

func (enum AutoscalerExpander) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown_expander"
	}
	return string(enum)
}

func (enum AutoscalerExpander) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *AutoscalerExpander) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = AutoscalerExpander(AutoscalerExpander(tmp).String())
	return nil
}

type CNI string

const (
	CNIUnknownCni = CNI("unknown_cni")
	// Cilium CNI will be configured (https://github.com/cilium/cilium)
	CNICilium = CNI("cilium")
	// Calico CNI will be configured (https://github.com/projectcalico/calico)
	CNICalico  = CNI("calico")
	CNIWeave   = CNI("weave")
	CNIFlannel = CNI("flannel")
	// Kilo CNI will be configured (https://github.com/squat/kilo/). Note that this CNI is only available for Kosmos clusters
	CNIKilo = CNI("kilo")
)

func (enum CNI) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown_cni"
	}
	return string(enum)
}

func (enum CNI) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *CNI) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = CNI(CNI(tmp).String())
	return nil
}

type ClusterStatus string

const (
	ClusterStatusUnknown = ClusterStatus("unknown")
	// Cluster is provisioning
	ClusterStatusCreating = ClusterStatus("creating")
	// Cluster is ready to use
	ClusterStatusReady = ClusterStatus("ready")
	// Cluster is waiting to be processed for deletion
	ClusterStatusDeleting = ClusterStatus("deleting")
	ClusterStatusDeleted  = ClusterStatus("deleted")
	// Cluster is updating its own configuration, it can be a version upgrade too
	ClusterStatusUpdating = ClusterStatus("updating")
	// Cluster is locked because an abuse has been detected or reported
	ClusterStatusLocked = ClusterStatus("locked")
	// Cluster has no associated pool
	ClusterStatusPoolRequired = ClusterStatus("pool_required")
)

func (enum ClusterStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum ClusterStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ClusterStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ClusterStatus(ClusterStatus(tmp).String())
	return nil
}

type ClusterTypeAvailability string

const (
	// Type is available in quantity
	ClusterTypeAvailabilityAvailable = ClusterTypeAvailability("available")
	// Limited availability
	ClusterTypeAvailabilityScarce = ClusterTypeAvailability("scarce")
	// Out of stock
	ClusterTypeAvailabilityShortage = ClusterTypeAvailability("shortage")
)

func (enum ClusterTypeAvailability) String() string {
	if enum == "" {
		// return default value if empty
		return "available"
	}
	return string(enum)
}

func (enum ClusterTypeAvailability) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ClusterTypeAvailability) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ClusterTypeAvailability(ClusterTypeAvailability(tmp).String())
	return nil
}

type ClusterTypeResiliency string

const (
	ClusterTypeResiliencyUnknownResiliency = ClusterTypeResiliency("unknown_resiliency")
	// The control plane is rescheduled on other machines in case of failure of a lower layer
	ClusterTypeResiliencyStandard = ClusterTypeResiliency("standard")
	// The control plane has replicas to ensure service continuity in case of failure of a lower layer.
	ClusterTypeResiliencyHighAvailability = ClusterTypeResiliency("high_availability")
)

func (enum ClusterTypeResiliency) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown_resiliency"
	}
	return string(enum)
}

func (enum ClusterTypeResiliency) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ClusterTypeResiliency) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ClusterTypeResiliency(ClusterTypeResiliency(tmp).String())
	return nil
}

type Ingress string

const (
	IngressUnknownIngress = Ingress("unknown_ingress")
	IngressNone           = Ingress("none")
	IngressNginx          = Ingress("nginx")
	IngressTraefik        = Ingress("traefik")
	IngressTraefik2       = Ingress("traefik2")
)

func (enum Ingress) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown_ingress"
	}
	return string(enum)
}

func (enum Ingress) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *Ingress) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = Ingress(Ingress(tmp).String())
	return nil
}

type ListClustersRequestOrderBy string

const (
	ListClustersRequestOrderByCreatedAtAsc  = ListClustersRequestOrderBy("created_at_asc")
	ListClustersRequestOrderByCreatedAtDesc = ListClustersRequestOrderBy("created_at_desc")
	ListClustersRequestOrderByUpdatedAtAsc  = ListClustersRequestOrderBy("updated_at_asc")
	ListClustersRequestOrderByUpdatedAtDesc = ListClustersRequestOrderBy("updated_at_desc")
	ListClustersRequestOrderByNameAsc       = ListClustersRequestOrderBy("name_asc")
	ListClustersRequestOrderByNameDesc      = ListClustersRequestOrderBy("name_desc")
	ListClustersRequestOrderByStatusAsc     = ListClustersRequestOrderBy("status_asc")
	ListClustersRequestOrderByStatusDesc    = ListClustersRequestOrderBy("status_desc")
	ListClustersRequestOrderByVersionAsc    = ListClustersRequestOrderBy("version_asc")
	ListClustersRequestOrderByVersionDesc   = ListClustersRequestOrderBy("version_desc")
)

func (enum ListClustersRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListClustersRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListClustersRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListClustersRequestOrderBy(ListClustersRequestOrderBy(tmp).String())
	return nil
}

type ListNodesRequestOrderBy string

const (
	ListNodesRequestOrderByCreatedAtAsc  = ListNodesRequestOrderBy("created_at_asc")
	ListNodesRequestOrderByCreatedAtDesc = ListNodesRequestOrderBy("created_at_desc")
)

func (enum ListNodesRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListNodesRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListNodesRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListNodesRequestOrderBy(ListNodesRequestOrderBy(tmp).String())
	return nil
}

type ListPoolsRequestOrderBy string

const (
	ListPoolsRequestOrderByCreatedAtAsc  = ListPoolsRequestOrderBy("created_at_asc")
	ListPoolsRequestOrderByCreatedAtDesc = ListPoolsRequestOrderBy("created_at_desc")
	ListPoolsRequestOrderByUpdatedAtAsc  = ListPoolsRequestOrderBy("updated_at_asc")
	ListPoolsRequestOrderByUpdatedAtDesc = ListPoolsRequestOrderBy("updated_at_desc")
	ListPoolsRequestOrderByNameAsc       = ListPoolsRequestOrderBy("name_asc")
	ListPoolsRequestOrderByNameDesc      = ListPoolsRequestOrderBy("name_desc")
	ListPoolsRequestOrderByStatusAsc     = ListPoolsRequestOrderBy("status_asc")
	ListPoolsRequestOrderByStatusDesc    = ListPoolsRequestOrderBy("status_desc")
	ListPoolsRequestOrderByVersionAsc    = ListPoolsRequestOrderBy("version_asc")
	ListPoolsRequestOrderByVersionDesc   = ListPoolsRequestOrderBy("version_desc")
)

func (enum ListPoolsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListPoolsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListPoolsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListPoolsRequestOrderBy(ListPoolsRequestOrderBy(tmp).String())
	return nil
}

type MaintenanceWindowDayOfTheWeek string

const (
	MaintenanceWindowDayOfTheWeekAny       = MaintenanceWindowDayOfTheWeek("any")
	MaintenanceWindowDayOfTheWeekMonday    = MaintenanceWindowDayOfTheWeek("monday")
	MaintenanceWindowDayOfTheWeekTuesday   = MaintenanceWindowDayOfTheWeek("tuesday")
	MaintenanceWindowDayOfTheWeekWednesday = MaintenanceWindowDayOfTheWeek("wednesday")
	MaintenanceWindowDayOfTheWeekThursday  = MaintenanceWindowDayOfTheWeek("thursday")
	MaintenanceWindowDayOfTheWeekFriday    = MaintenanceWindowDayOfTheWeek("friday")
	MaintenanceWindowDayOfTheWeekSaturday  = MaintenanceWindowDayOfTheWeek("saturday")
	MaintenanceWindowDayOfTheWeekSunday    = MaintenanceWindowDayOfTheWeek("sunday")
)

func (enum MaintenanceWindowDayOfTheWeek) String() string {
	if enum == "" {
		// return default value if empty
		return "any"
	}
	return string(enum)
}

func (enum MaintenanceWindowDayOfTheWeek) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *MaintenanceWindowDayOfTheWeek) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = MaintenanceWindowDayOfTheWeek(MaintenanceWindowDayOfTheWeek(tmp).String())
	return nil
}

type NodeStatus string

const (
	NodeStatusUnknown = NodeStatus("unknown")
	// Node is provisioning
	NodeStatusCreating = NodeStatus("creating")
	// Node is unable to connect to apiserver
	NodeStatusNotReady = NodeStatus("not_ready")
	// Node is ready to execute workload (marked schedulable by k8s scheduler)
	NodeStatusReady = NodeStatus("ready")
	// Node is waiting to be processed for deletion
	NodeStatusDeleting = NodeStatus("deleting")
	NodeStatusDeleted  = NodeStatus("deleted")
	// Node is locked because an abuse has been detected or reported
	NodeStatusLocked = NodeStatus("locked")
	// Node is rebooting
	NodeStatusRebooting     = NodeStatus("rebooting")
	NodeStatusCreationError = NodeStatus("creation_error")
	NodeStatusUpgrading     = NodeStatus("upgrading")
	NodeStatusStarting      = NodeStatus("starting")
	NodeStatusRegistering   = NodeStatus("registering")
)

func (enum NodeStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum NodeStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *NodeStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = NodeStatus(NodeStatus(tmp).String())
	return nil
}

type PoolStatus string

const (
	PoolStatusUnknown = PoolStatus("unknown")
	// Pool has the right amount of nodes and is ready to process the workload
	PoolStatusReady = PoolStatus("ready")
	// Pool is waiting to be processed for deletion
	PoolStatusDeleting = PoolStatus("deleting")
	PoolStatusDeleted  = PoolStatus("deleted")
	// Pool is growing or shrinking
	PoolStatusScaling = PoolStatus("scaling")
	PoolStatusWarning = PoolStatus("warning")
	// Pool is locked because an abuse has been detected or reported
	PoolStatusLocked = PoolStatus("locked")
	// Pool is upgrading its Kubernetes version
	PoolStatusUpgrading = PoolStatus("upgrading")
)

func (enum PoolStatus) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown"
	}
	return string(enum)
}

func (enum PoolStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *PoolStatus) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = PoolStatus(PoolStatus(tmp).String())
	return nil
}

type PoolVolumeType string

const (
	PoolVolumeTypeDefaultVolumeType = PoolVolumeType("default_volume_type")
	// Local Block Storage: your system is stored locally on your node hypervisor. Lower latency, no persistence across node replacements
	PoolVolumeTypeLSSD = PoolVolumeType("l_ssd")
	// Remote Block Storage: your system is stored on a centralized and resilient cluster. Higher latency, persistence across node replacements
	PoolVolumeTypeBSSD = PoolVolumeType("b_ssd")
)

func (enum PoolVolumeType) String() string {
	if enum == "" {
		// return default value if empty
		return "default_volume_type"
	}
	return string(enum)
}

func (enum PoolVolumeType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *PoolVolumeType) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = PoolVolumeType(PoolVolumeType(tmp).String())
	return nil
}

type Runtime string

const (
	RuntimeUnknownRuntime = Runtime("unknown_runtime")
	RuntimeDocker         = Runtime("docker")
	// Containerd Runtime will be configured (https://github.com/containerd/containerd)
	RuntimeContainerd = Runtime("containerd")
	RuntimeCrio       = Runtime("crio")
)

func (enum Runtime) String() string {
	if enum == "" {
		// return default value if empty
		return "unknown_runtime"
	}
	return string(enum)
}

func (enum Runtime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *Runtime) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = Runtime(Runtime(tmp).String())
	return nil
}

// Cluster: cluster.
type Cluster struct {
	// ID: cluster ID.
	ID string `json:"id"`
	// Type: cluster type.
	Type string `json:"type"`
	// Name: cluster name.
	Name string `json:"name"`
	// Status: status of the cluster.
	// Default value: unknown
	Status ClusterStatus `json:"status"`
	// Version: kubernetes version of the cluster.
	Version string `json:"version"`
	// Region: region in which the cluster is deployed.
	Region scw.Region `json:"region"`
	// OrganizationID: ID of the Organization owning the cluster.
	OrganizationID string `json:"organization_id"`
	// ProjectID: ID of the Project owning the cluster.
	ProjectID string `json:"project_id"`
	// Tags: tags associated with the cluster.
	Tags []string `json:"tags"`
	// Cni: container Network Interface (CNI) plugin running in the cluster.
	// Default value: unknown_cni
	Cni CNI `json:"cni"`
	// Description: cluster description.
	Description string `json:"description"`
	// ClusterURL: kubernetes API server URL of the cluster.
	ClusterURL string `json:"cluster_url"`
	// DNSWildcard: wildcard DNS resolving all the ready cluster nodes.
	DNSWildcard string `json:"dns_wildcard"`
	// CreatedAt: date on which the cluster was created.
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: date on which the cluster was last updated.
	UpdatedAt *time.Time `json:"updated_at"`
	// AutoscalerConfig: autoscaler config for the cluster.
	AutoscalerConfig *ClusterAutoscalerConfig `json:"autoscaler_config"`
	// Deprecated: DashboardEnabled: defines whether the Kubernetes dashboard is enabled for the cluster.
	DashboardEnabled *bool `json:"dashboard_enabled,omitempty"`
	// Deprecated: Ingress: managed Ingress controller used in the cluster (deprecated feature).
	// Default value: unknown_ingress
	Ingress *Ingress `json:"ingress,omitempty"`
	// AutoUpgrade: auto upgrade configuration of the cluster.
	AutoUpgrade *ClusterAutoUpgrade `json:"auto_upgrade"`
	// UpgradeAvailable: defines whether a new Kubernetes version is available.
	UpgradeAvailable bool `json:"upgrade_available"`
	// FeatureGates: list of enabled feature gates.
	FeatureGates []string `json:"feature_gates"`
	// AdmissionPlugins: list of enabled admission plugins.
	AdmissionPlugins []string `json:"admission_plugins"`
	// OpenIDConnectConfig: this configuration enables to update the OpenID Connect configuration of the Kubernetes API server.
	OpenIDConnectConfig *ClusterOpenIDConnectConfig `json:"open_id_connect_config"`
	// ApiserverCertSans: additional Subject Alternative Names for the Kubernetes API server certificate.
	ApiserverCertSans []string `json:"apiserver_cert_sans"`
	// PrivateNetworkID: private network ID for internal cluster communication.
	PrivateNetworkID *string `json:"private_network_id"`
	// CommitmentEndsAt: date on which it will be possible to switch to a smaller offer.
	CommitmentEndsAt *time.Time `json:"commitment_ends_at"`
}

// ClusterAutoUpgrade: cluster. auto upgrade.
type ClusterAutoUpgrade struct {
	// Enabled: defines whether auto upgrade is enabled for the cluster.
	Enabled bool `json:"enabled"`
	// MaintenanceWindow: maintenance window of the cluster auto upgrades.
	MaintenanceWindow *MaintenanceWindow `json:"maintenance_window"`
}

// ClusterAutoscalerConfig: cluster. autoscaler config.
type ClusterAutoscalerConfig struct {
	// ScaleDownDisabled: disable the cluster autoscaler.
	ScaleDownDisabled bool `json:"scale_down_disabled"`
	// ScaleDownDelayAfterAdd: how long after scale up that scale down evaluation resumes.
	ScaleDownDelayAfterAdd string `json:"scale_down_delay_after_add"`
	// Estimator: type of resource estimator to be used in scale up.
	// Default value: unknown_estimator
	Estimator AutoscalerEstimator `json:"estimator"`
	// Expander: type of node group expander to be used in scale up.
	// Default value: unknown_expander
	Expander AutoscalerExpander `json:"expander"`
	// IgnoreDaemonsetsUtilization: ignore DaemonSet pods when calculating resource utilization for scaling down.
	IgnoreDaemonsetsUtilization bool `json:"ignore_daemonsets_utilization"`
	// BalanceSimilarNodeGroups: detect similar node groups and balance the number of nodes between them.
	BalanceSimilarNodeGroups bool `json:"balance_similar_node_groups"`
	// ExpendablePodsPriorityCutoff: pods with priority below cutoff will be expendable. They can be killed without any consideration during scale down and they won't cause scale up. Pods with null priority (PodPriority disabled) are non expendable.
	ExpendablePodsPriorityCutoff int32 `json:"expendable_pods_priority_cutoff"`
	// ScaleDownUnneededTime: how long a node should be unneeded before it is eligible to be scaled down.
	ScaleDownUnneededTime string `json:"scale_down_unneeded_time"`
	// ScaleDownUtilizationThreshold: node utilization level, defined as a sum of requested resources divided by capacity, below which a node can be considered for scale down.
	ScaleDownUtilizationThreshold float32 `json:"scale_down_utilization_threshold"`
	// MaxGracefulTerminationSec: maximum number of seconds the cluster autoscaler waits for pod termination when trying to scale down a node.
	MaxGracefulTerminationSec uint32 `json:"max_graceful_termination_sec"`
}

// ClusterOpenIDConnectConfig: cluster. open id connect config.
type ClusterOpenIDConnectConfig struct {
	// IssuerURL: URL of the provider which allows the API server to discover public signing keys. Only URLs using the `https://` scheme are accepted. This is typically the provider's discovery URL without a path, for example "https://accounts.google.com" or "https://login.salesforce.com".
	IssuerURL string `json:"issuer_url"`
	// ClientID: a client ID that all tokens must be issued for.
	ClientID string `json:"client_id"`
	// UsernameClaim: jWT claim to use as the user name. The default is `sub`, which is expected to be the end user's unique identifier. Admins can choose other claims, such as `email` or `name`, depending on their provider. However, claims other than `email` will be prefixed with the issuer URL to prevent name collision.
	UsernameClaim string `json:"username_claim"`
	// UsernamePrefix: prefix prepended to username claims to prevent name collision (such as `system:` users). For example, the value `oidc:` will create usernames like `oidc:jane.doe`. If this flag is not provided and `username_claim` is a value other than `email`, the prefix defaults to `( Issuer URL )#` where `( Issuer URL )` is the value of `issuer_url`. The value `-` can be used to disable all prefixing.
	UsernamePrefix string `json:"username_prefix"`
	// GroupsClaim: jWT claim to use as the user's group.
	GroupsClaim []string `json:"groups_claim"`
	// GroupsPrefix: prefix prepended to group claims to prevent name collision (such as `system:` groups). For example, the value `oidc:` will create group names like `oidc:engineering` and `oidc:infra`.
	GroupsPrefix string `json:"groups_prefix"`
	// RequiredClaim: multiple key=value pairs describing a required claim in the ID token. If set, the claims are verified to be present in the ID token with a matching value.
	RequiredClaim []string `json:"required_claim"`
}

// ClusterType: cluster type.
type ClusterType struct {
	// Name: cluster type name.
	Name string `json:"name"`
	// Availability: cluster type availability.
	// Default value: available
	Availability ClusterTypeAvailability `json:"availability"`
	// MaxNodes: maximum number of nodes supported by the offer.
	MaxNodes uint32 `json:"max_nodes"`
	// CommitmentDelay: time period during which you can no longer switch to a lower offer.
	CommitmentDelay *scw.Duration `json:"commitment_delay"`
	// SLA: value of the Service Level Agreement of the offer.
	SLA float32 `json:"sla"`
	// Resiliency: resiliency offered by the offer.
	// Default value: unknown_resiliency
	Resiliency ClusterTypeResiliency `json:"resiliency"`
	// Memory: max RAM allowed for the control plane.
	Memory scw.Size `json:"memory"`
	// Dedicated: returns information if this offer uses dedicated resources.
	Dedicated bool `json:"dedicated"`
}

// CreateClusterRequestAutoUpgrade: create cluster request. auto upgrade.
type CreateClusterRequestAutoUpgrade struct {
	// Enable: defines whether auto upgrade is enabled for the cluster.
	Enable bool `json:"enable"`
	// MaintenanceWindow: maintenance window of the cluster auto upgrades.
	MaintenanceWindow *MaintenanceWindow `json:"maintenance_window"`
}

// CreateClusterRequestAutoscalerConfig: create cluster request. autoscaler config.
type CreateClusterRequestAutoscalerConfig struct {
	// ScaleDownDisabled: disable the cluster autoscaler.
	ScaleDownDisabled *bool `json:"scale_down_disabled"`
	// ScaleDownDelayAfterAdd: how long after scale up that scale down evaluation resumes.
	ScaleDownDelayAfterAdd *string `json:"scale_down_delay_after_add"`
	// Estimator: type of resource estimator to be used in scale up.
	// Default value: unknown_estimator
	Estimator AutoscalerEstimator `json:"estimator"`
	// Expander: type of node group expander to be used in scale up.
	// Default value: unknown_expander
	Expander AutoscalerExpander `json:"expander"`
	// IgnoreDaemonsetsUtilization: ignore DaemonSet pods when calculating resource utilization for scaling down.
	IgnoreDaemonsetsUtilization *bool `json:"ignore_daemonsets_utilization"`
	// BalanceSimilarNodeGroups: detect similar node groups and balance the number of nodes between them.
	BalanceSimilarNodeGroups *bool `json:"balance_similar_node_groups"`
	// ExpendablePodsPriorityCutoff: pods with priority below cutoff will be expendable. They can be killed without any consideration during scale down and they won't cause scale up. Pods with null priority (PodPriority disabled) are non expendable.
	ExpendablePodsPriorityCutoff *int32 `json:"expendable_pods_priority_cutoff"`
	// ScaleDownUnneededTime: how long a node should be unneeded before it is eligible to be scaled down.
	ScaleDownUnneededTime *string `json:"scale_down_unneeded_time"`
	// ScaleDownUtilizationThreshold: node utilization level, defined as a sum of requested resources divided by capacity, below which a node can be considered for scale down.
	ScaleDownUtilizationThreshold *float32 `json:"scale_down_utilization_threshold"`
	// MaxGracefulTerminationSec: maximum number of seconds the cluster autoscaler waits for pod termination when trying to scale down a node.
	MaxGracefulTerminationSec *uint32 `json:"max_graceful_termination_sec"`
}

// CreateClusterRequestOpenIDConnectConfig: create cluster request. open id connect config.
type CreateClusterRequestOpenIDConnectConfig struct {
	// IssuerURL: URL of the provider which allows the API server to discover public signing keys. Only URLs using the `https://` scheme are accepted. This is typically the provider's discovery URL without a path, for example "https://accounts.google.com" or "https://login.salesforce.com".
	IssuerURL string `json:"issuer_url"`
	// ClientID: a client ID that all tokens must be issued for.
	ClientID string `json:"client_id"`
	// UsernameClaim: jWT claim to use as the user name. The default is `sub`, which is expected to be the end user's unique identifier. Admins can choose other claims, such as `email` or `name`, depending on their provider. However, claims other than `email` will be prefixed with the issuer URL to prevent name collision.
	UsernameClaim *string `json:"username_claim"`
	// UsernamePrefix: prefix prepended to username claims to prevent name collision (such as `system:` users). For example, the value `oidc:` will create usernames like `oidc:jane.doe`. If this flag is not provided and `username_claim` is a value other than `email`, the prefix defaults to `( Issuer URL )#` where `( Issuer URL )` is the value of `issuer_url`. The value `-` can be used to disable all prefixing.
	UsernamePrefix *string `json:"username_prefix"`
	// GroupsClaim: jWT claim to use as the user's group.
	GroupsClaim *[]string `json:"groups_claim"`
	// GroupsPrefix: prefix prepended to group claims to prevent name collision (such as `system:` groups). For example, the value `oidc:` will create group names like `oidc:engineering` and `oidc:infra`.
	GroupsPrefix *string `json:"groups_prefix"`
	// RequiredClaim: multiple key=value pairs describing a required claim in the ID token. If set, the claims are verified to be present in the ID token with a matching value.
	RequiredClaim *[]string `json:"required_claim"`
}

// CreateClusterRequestPoolConfig: create cluster request. pool config.
type CreateClusterRequestPoolConfig struct {
	// Name: name of the pool.
	Name string `json:"name"`
	// NodeType: node type is the type of Scaleway Instance wanted for the pool. Nodes with insufficient memory are not eligible (DEV1-S, PLAY2-PICO, STARDUST). 'external' is a special node type used to provision instances from other cloud providers in a Kosmos Cluster.
	NodeType string `json:"node_type"`
	// PlacementGroupID: placement group ID in which all the nodes of the pool will be created.
	PlacementGroupID *string `json:"placement_group_id"`
	// Autoscaling: defines whether the autoscaling feature is enabled for the pool.
	Autoscaling bool `json:"autoscaling"`
	// Size: size (number of nodes) of the pool.
	Size uint32 `json:"size"`
	// MinSize: defines the minimum size of the pool. Note that this field is only used when autoscaling is enabled on the pool.
	MinSize *uint32 `json:"min_size"`
	// MaxSize: defines the maximum size of the pool. Note that this field is only used when autoscaling is enabled on the pool.
	MaxSize *uint32 `json:"max_size"`
	// ContainerRuntime: customization of the container runtime is available for each pool. Note that `docker` has been deprecated since version 1.20 and will be removed by version 1.24.
	// Default value: unknown_runtime
	ContainerRuntime Runtime `json:"container_runtime"`
	// Autohealing: defines whether the autohealing feature is enabled for the pool.
	Autohealing bool `json:"autohealing"`
	// Tags: tags associated with the pool.
	Tags []string `json:"tags"`
	// KubeletArgs: kubelet arguments to be used by this pool. Note that this feature is experimental.
	KubeletArgs map[string]string `json:"kubelet_args"`
	// UpgradePolicy: pool upgrade policy.
	UpgradePolicy *CreateClusterRequestPoolConfigUpgradePolicy `json:"upgrade_policy"`
	// Zone: zone in which the pool's nodes will be spawned.
	Zone scw.Zone `json:"zone"`
	// RootVolumeType: defines the system volume disk type. Two different types of volume (`volume_type`) are provided: `l_ssd` is a local block storage which means your system is stored locally on your node's hypervisor. `b_ssd` is a remote block storage which means your system is stored on a centralized and resilient cluster.
	// Default value: default_volume_type
	RootVolumeType PoolVolumeType `json:"root_volume_type"`
	// RootVolumeSize: system volume disk size.
	RootVolumeSize *scw.Size `json:"root_volume_size"`
}

// CreateClusterRequestPoolConfigUpgradePolicy: create cluster request. pool config. upgrade policy.
type CreateClusterRequestPoolConfigUpgradePolicy struct {
	// MaxUnavailable: the maximum number of nodes that can be not ready at the same time.
	MaxUnavailable *uint32 `json:"max_unavailable"`
	// MaxSurge: the maximum number of nodes to be created during the upgrade.
	MaxSurge *uint32 `json:"max_surge"`
}

type CreatePoolRequestUpgradePolicy struct {
	MaxUnavailable *uint32 `json:"max_unavailable"`

	MaxSurge *uint32 `json:"max_surge"`
}

type ExternalNode struct {
	ID string `json:"id"`

	Name string `json:"name"`

	ClusterURL string `json:"cluster_url"`

	PoolVersion string `json:"pool_version"`

	ClusterCa string `json:"cluster_ca"`

	KubeToken string `json:"kube_token"`

	KubeletConfig string `json:"kubelet_config"`

	ExternalIP string `json:"external_ip"`
}

// ListClusterAvailableTypesResponse: list cluster available types response.
type ListClusterAvailableTypesResponse struct {
	// ClusterTypes: available cluster types for the cluster.
	ClusterTypes []*ClusterType `json:"cluster_types"`
	// TotalCount: total number of types.
	TotalCount uint32 `json:"total_count"`
}

// ListClusterAvailableVersionsResponse: list cluster available versions response.
type ListClusterAvailableVersionsResponse struct {
	// Versions: available Kubernetes versions for the cluster.
	Versions []*Version `json:"versions"`
}

// ListClusterTypesResponse: list cluster types response.
type ListClusterTypesResponse struct {
	// TotalCount: total number of cluster-types.
	TotalCount uint32 `json:"total_count"`
	// ClusterTypes: paginated returned cluster-types.
	ClusterTypes []*ClusterType `json:"cluster_types"`
}

// ListClustersResponse: list clusters response.
type ListClustersResponse struct {
	// TotalCount: total number of clusters.
	TotalCount uint32 `json:"total_count"`
	// Clusters: paginated returned clusters.
	Clusters []*Cluster `json:"clusters"`
}

// ListNodesResponse: list nodes response.
type ListNodesResponse struct {
	// TotalCount: total number of nodes.
	TotalCount uint32 `json:"total_count"`
	// Nodes: paginated returned nodes.
	Nodes []*Node `json:"nodes"`
}

// ListPoolsResponse: list pools response.
type ListPoolsResponse struct {
	// TotalCount: total number of pools that exists for the cluster.
	TotalCount uint32 `json:"total_count"`
	// Pools: paginated returned pools.
	Pools []*Pool `json:"pools"`
}

// ListVersionsResponse: list versions response.
type ListVersionsResponse struct {
	// Versions: available Kubernetes versions.
	Versions []*Version `json:"versions"`
}

// MaintenanceWindow: maintenance window.
type MaintenanceWindow struct {
	// StartHour: start time of the two-hour maintenance window.
	StartHour uint32 `json:"start_hour"`
	// Day: day of the week for the maintenance window.
	// Default value: any
	Day MaintenanceWindowDayOfTheWeek `json:"day"`
}

// Node: node.
type Node struct {
	// ID: node ID.
	ID string `json:"id"`
	// PoolID: pool ID of the node.
	PoolID string `json:"pool_id"`
	// ClusterID: cluster ID of the node.
	ClusterID string `json:"cluster_id"`
	// ProviderID: underlying instance ID. It is prefixed by instance type and location information (see https://pkg.go.dev/k8s.io/api/core/v1#NodeSpec.ProviderID).
	ProviderID string `json:"provider_id"`
	// Region: cluster region of the node.
	Region scw.Region `json:"region"`
	// Name: name of the node.
	Name string `json:"name"`
	// Deprecated: PublicIPV4: public IPv4 address of the node.
	PublicIPV4 *net.IP `json:"public_ip_v4,omitempty"`
	// Deprecated: PublicIPV6: public IPv6 address of the node.
	PublicIPV6 *net.IP `json:"public_ip_v6,omitempty"`
	// Deprecated: Conditions: conditions of the node. These conditions contain the Node Problem Detector conditions, as well as some in house conditions.
	Conditions *map[string]string `json:"conditions,omitempty"`
	// Status: status of the node.
	// Default value: unknown
	Status NodeStatus `json:"status"`
	// ErrorMessage: details of the error, if any occurred when managing the node.
	ErrorMessage *string `json:"error_message"`
	// CreatedAt: date on which the node was created.
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: date on which the node was last updated.
	UpdatedAt *time.Time `json:"updated_at"`
}

// Pool: pool.
type Pool struct {
	// ID: pool ID.
	ID string `json:"id"`
	// ClusterID: cluster ID of the pool.
	ClusterID string `json:"cluster_id"`
	// CreatedAt: date on which the pool was created.
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: date on which the pool was last updated.
	UpdatedAt *time.Time `json:"updated_at"`
	// Name: pool name.
	Name string `json:"name"`
	// Status: pool status.
	// Default value: unknown
	Status PoolStatus `json:"status"`
	// Version: pool version.
	Version string `json:"version"`
	// NodeType: node type is the type of Scaleway Instance wanted for the pool. Nodes with insufficient memory are not eligible (DEV1-S, PLAY2-PICO, STARDUST). 'external' is a special node type used to provision instances from other cloud providers in a Kosmos Cluster.
	NodeType string `json:"node_type"`
	// Autoscaling: defines whether the autoscaling feature is enabled for the pool.
	Autoscaling bool `json:"autoscaling"`
	// Size: size (number of nodes) of the pool.
	Size uint32 `json:"size"`
	// MinSize: defines the minimum size of the pool. Note that this field is only used when autoscaling is enabled on the pool.
	MinSize uint32 `json:"min_size"`
	// MaxSize: defines the maximum size of the pool. Note that this field is only used when autoscaling is enabled on the pool.
	MaxSize uint32 `json:"max_size"`
	// ContainerRuntime: customization of the container runtime is available for each pool. Note that `docker` has been deprecated since version 1.20 and will be removed by version 1.24.
	// Default value: unknown_runtime
	ContainerRuntime Runtime `json:"container_runtime"`
	// Autohealing: defines whether the autohealing feature is enabled for the pool.
	Autohealing bool `json:"autohealing"`
	// Tags: tags associated with the pool.
	Tags []string `json:"tags"`
	// PlacementGroupID: placement group ID in which all the nodes of the pool will be created.
	PlacementGroupID *string `json:"placement_group_id"`
	// KubeletArgs: kubelet arguments to be used by this pool. Note that this feature is experimental.
	KubeletArgs map[string]string `json:"kubelet_args"`
	// UpgradePolicy: pool upgrade policy.
	UpgradePolicy *PoolUpgradePolicy `json:"upgrade_policy"`
	// Zone: zone in which the pool's nodes will be spawned.
	Zone scw.Zone `json:"zone"`
	// RootVolumeType: defines the system volume disk type. Two different types of volume (`volume_type`) are provided: `l_ssd` is a local block storage which means your system is stored locally on your node's hypervisor. `b_ssd` is a remote block storage which means your system is stored on a centralized and resilient cluster.
	// Default value: default_volume_type
	RootVolumeType PoolVolumeType `json:"root_volume_type"`
	// RootVolumeSize: system volume disk size.
	RootVolumeSize *scw.Size `json:"root_volume_size"`
	// Region: cluster region of the pool.
	Region scw.Region `json:"region"`
}

type PoolUpgradePolicy struct {
	MaxUnavailable uint32 `json:"max_unavailable"`

	MaxSurge uint32 `json:"max_surge"`
}

// UpdateClusterRequestAutoUpgrade: update cluster request. auto upgrade.
type UpdateClusterRequestAutoUpgrade struct {
	// Enable: defines whether auto upgrade is enabled for the cluster.
	Enable *bool `json:"enable"`
	// MaintenanceWindow: maintenance window of the cluster auto upgrades.
	MaintenanceWindow *MaintenanceWindow `json:"maintenance_window"`
}

// UpdateClusterRequestAutoscalerConfig: update cluster request. autoscaler config.
type UpdateClusterRequestAutoscalerConfig struct {
	// ScaleDownDisabled: disable the cluster autoscaler.
	ScaleDownDisabled *bool `json:"scale_down_disabled"`
	// ScaleDownDelayAfterAdd: how long after scale up that scale down evaluation resumes.
	ScaleDownDelayAfterAdd *string `json:"scale_down_delay_after_add"`
	// Estimator: type of resource estimator to be used in scale up.
	// Default value: unknown_estimator
	Estimator AutoscalerEstimator `json:"estimator"`
	// Expander: type of node group expander to be used in scale up.
	// Default value: unknown_expander
	Expander AutoscalerExpander `json:"expander"`
	// IgnoreDaemonsetsUtilization: ignore DaemonSet pods when calculating resource utilization for scaling down.
	IgnoreDaemonsetsUtilization *bool `json:"ignore_daemonsets_utilization"`
	// BalanceSimilarNodeGroups: detect similar node groups and balance the number of nodes between them.
	BalanceSimilarNodeGroups *bool `json:"balance_similar_node_groups"`
	// ExpendablePodsPriorityCutoff: pods with priority below cutoff will be expendable. They can be killed without any consideration during scale down and they won't cause scale up. Pods with null priority (PodPriority disabled) are non expendable.
	ExpendablePodsPriorityCutoff *int32 `json:"expendable_pods_priority_cutoff"`
	// ScaleDownUnneededTime: how long a node should be unneeded before it is eligible to be scaled down.
	ScaleDownUnneededTime *string `json:"scale_down_unneeded_time"`
	// ScaleDownUtilizationThreshold: node utilization level, defined as a sum of requested resources divided by capacity, below which a node can be considered for scale down.
	ScaleDownUtilizationThreshold *float32 `json:"scale_down_utilization_threshold"`
	// MaxGracefulTerminationSec: maximum number of seconds the cluster autoscaler waits for pod termination when trying to scale down a node.
	MaxGracefulTerminationSec *uint32 `json:"max_graceful_termination_sec"`
}

// UpdateClusterRequestOpenIDConnectConfig: update cluster request. open id connect config.
type UpdateClusterRequestOpenIDConnectConfig struct {
	// IssuerURL: URL of the provider which allows the API server to discover public signing keys. Only URLs using the `https://` scheme are accepted. This is typically the provider's discovery URL without a path, for example "https://accounts.google.com" or "https://login.salesforce.com".
	IssuerURL *string `json:"issuer_url"`
	// ClientID: a client ID that all tokens must be issued for.
	ClientID *string `json:"client_id"`
	// UsernameClaim: jWT claim to use as the user name. The default is `sub`, which is expected to be the end user's unique identifier. Admins can choose other claims, such as `email` or `name`, depending on their provider. However, claims other than `email` will be prefixed with the issuer URL to prevent name collision.
	UsernameClaim *string `json:"username_claim"`
	// UsernamePrefix: prefix prepended to username claims to prevent name collision (such as `system:` users). For example, the value `oidc:` will create usernames like `oidc:jane.doe`. If this flag is not provided and `username_claim` is a value other than `email`, the prefix defaults to `( Issuer URL )#` where `( Issuer URL )` is the value of `issuer_url`. The value `-` can be used to disable all prefixing.
	UsernamePrefix *string `json:"username_prefix"`
	// GroupsClaim: jWT claim to use as the user's group.
	GroupsClaim *[]string `json:"groups_claim"`
	// GroupsPrefix: prefix prepended to group claims to prevent name collision (such as `system:` groups). For example, the value `oidc:` will create group names like `oidc:engineering` and `oidc:infra`.
	GroupsPrefix *string `json:"groups_prefix"`
	// RequiredClaim: multiple key=value pairs describing a required claim in the ID token. If set, the claims are verified to be present in the ID token with a matching value.
	RequiredClaim *[]string `json:"required_claim"`
}

type UpdatePoolRequestUpgradePolicy struct {
	MaxUnavailable *uint32 `json:"max_unavailable"`

	MaxSurge *uint32 `json:"max_surge"`
}

// Version: version.
type Version struct {
	// Name: name of the Kubernetes version.
	Name string `json:"name"`
	// Label: label of the Kubernetes version.
	Label string `json:"label"`
	// Region: region in which this version is available.
	Region scw.Region `json:"region"`
	// AvailableCnis: supported Container Network Interface (CNI) plugins for this version.
	AvailableCnis []CNI `json:"available_cnis"`
	// Deprecated: AvailableIngresses: supported Ingress Controllers for this version.
	AvailableIngresses *[]Ingress `json:"available_ingresses,omitempty"`
	// AvailableContainerRuntimes: supported container runtimes for this version.
	AvailableContainerRuntimes []Runtime `json:"available_container_runtimes"`
	// AvailableFeatureGates: supported feature gates for this version.
	AvailableFeatureGates []string `json:"available_feature_gates"`
	// AvailableAdmissionPlugins: supported admission plugins for this version.
	AvailableAdmissionPlugins []string `json:"available_admission_plugins"`
	// AvailableKubeletArgs: supported kubelet arguments for this version.
	AvailableKubeletArgs map[string]string `json:"available_kubelet_args"`
}

// Service API

// Regions list localities the api is available in
func (s *API) Regions() []scw.Region {
	return []scw.Region{scw.RegionFrPar, scw.RegionNlAms, scw.RegionPlWaw}
}

type ListClustersRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// OrganizationID: organization ID on which to filter the returned clusters.
	OrganizationID *string `json:"-"`
	// ProjectID: project ID on which to filter the returned clusters.
	ProjectID *string `json:"-"`
	// OrderBy: sort order of returned clusters.
	// Default value: created_at_asc
	OrderBy ListClustersRequestOrderBy `json:"-"`
	// Page: page number to return for clusters, from the paginated results.
	Page *int32 `json:"-"`
	// PageSize: maximum number of clusters per page.
	PageSize *uint32 `json:"-"`
	// Name: name to filter on, only clusters containing this substring in their name will be returned.
	Name *string `json:"-"`
	// Status: status to filter on, only clusters with this status will be returned.
	// Default value: unknown
	Status ClusterStatus `json:"-"`
	// Type: type to filter on, only clusters with this type will be returned.
	Type *string `json:"-"`
}

// ListClusters: list Clusters.
// List all existing Kubernetes clusters in a specific region.
func (s *API) ListClusters(req *ListClustersRequest, opts ...scw.RequestOption) (*ListClustersResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "project_id", req.ProjectID)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "status", req.Status)
	parameter.AddToQuery(query, "type", req.Type)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/clusters",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListClustersResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateClusterRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// Deprecated: OrganizationID: organization ID in which the cluster will be created.
	// Precisely one of OrganizationID, ProjectID must be set.
	OrganizationID *string `json:"organization_id,omitempty"`
	// ProjectID: project ID in which the cluster will be created.
	// Precisely one of OrganizationID, ProjectID must be set.
	ProjectID *string `json:"project_id,omitempty"`
	// Type: type of the cluster (possible values are kapsule, multicloud, kapsule-dedicated-8, kapsule-dedicated-16).
	Type string `json:"type"`
	// Name: cluster name.
	Name string `json:"name"`
	// Description: cluster description.
	Description string `json:"description"`
	// Tags: tags associated with the cluster.
	Tags []string `json:"tags"`
	// Version: kubernetes version of the cluster.
	Version string `json:"version"`
	// Cni: container Network Interface (CNI) plugin running in the cluster.
	// Default value: unknown_cni
	Cni CNI `json:"cni"`
	// Deprecated: EnableDashboard: defines whether the Kubernetes Dashboard is enabled in the cluster.
	EnableDashboard *bool `json:"enable_dashboard,omitempty"`
	// Deprecated: Ingress: ingress Controller running in the cluster (deprecated feature).
	// Default value: unknown_ingress
	Ingress *Ingress `json:"ingress,omitempty"`
	// Pools: pools created along with the cluster.
	Pools []*CreateClusterRequestPoolConfig `json:"pools"`
	// AutoscalerConfig: autoscaler configuration for the cluster. It allows you to set (to an extent) your preferred autoscaler configuration, which is an implementation of the cluster-autoscaler (https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/).
	AutoscalerConfig *CreateClusterRequestAutoscalerConfig `json:"autoscaler_config"`
	// AutoUpgrade: auto upgrade configuration of the cluster. This configuration enables to set a specific 2-hour time window in which the cluster can be automatically updated to the latest patch version.
	AutoUpgrade *CreateClusterRequestAutoUpgrade `json:"auto_upgrade"`
	// FeatureGates: list of feature gates to enable.
	FeatureGates []string `json:"feature_gates"`
	// AdmissionPlugins: list of admission plugins to enable.
	AdmissionPlugins []string `json:"admission_plugins"`
	// OpenIDConnectConfig: openID Connect configuration of the cluster. This configuration enables to update the OpenID Connect configuration of the Kubernetes API server.
	OpenIDConnectConfig *CreateClusterRequestOpenIDConnectConfig `json:"open_id_connect_config"`
	// ApiserverCertSans: additional Subject Alternative Names for the Kubernetes API server certificate.
	ApiserverCertSans []string `json:"apiserver_cert_sans"`
	// PrivateNetworkID: private network ID for internal cluster communication (cannot be changed later).
	PrivateNetworkID *string `json:"private_network_id"`
}

// CreateCluster: create a new Cluster.
// Create a new Kubernetes cluster in a Scaleway region.
func (s *API) CreateCluster(req *CreateClusterRequest, opts ...scw.RequestOption) (*Cluster, error) {
	var err error

	defaultProjectID, exist := s.client.GetDefaultProjectID()
	if exist && req.OrganizationID == nil && req.ProjectID == nil {
		req.ProjectID = &defaultProjectID
	}

	defaultOrganizationID, exist := s.client.GetDefaultOrganizationID()
	if exist && req.OrganizationID == nil && req.ProjectID == nil {
		req.OrganizationID = &defaultOrganizationID
	}

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("k8s")
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/clusters",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Cluster

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetClusterRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// ClusterID: ID of the requested cluster.
	ClusterID string `json:"-"`
}

// GetCluster: get a Cluster.
// Retrieve information about a specific Kubernetes cluster.
func (s *API) GetCluster(req *GetClusterRequest, opts ...scw.RequestOption) (*Cluster, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "",
		Headers: http.Header{},
	}

	var resp Cluster

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateClusterRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// ClusterID: ID of the cluster to update.
	ClusterID string `json:"-"`
	// Name: new external name for the cluster.
	Name *string `json:"name"`
	// Description: new description for the cluster.
	Description *string `json:"description"`
	// Tags: new tags associated with the cluster.
	Tags *[]string `json:"tags"`
	// AutoscalerConfig: new autoscaler config for the cluster.
	AutoscalerConfig *UpdateClusterRequestAutoscalerConfig `json:"autoscaler_config"`
	// Deprecated: EnableDashboard: new value for the Kubernetes Dashboard enablement.
	EnableDashboard *bool `json:"enable_dashboard,omitempty"`
	// Deprecated: Ingress: new Ingress Controller for the cluster (deprecated feature).
	// Default value: unknown_ingress
	Ingress *Ingress `json:"ingress,omitempty"`
	// AutoUpgrade: new auto upgrade configuration for the cluster. Note that all fields need to be set.
	AutoUpgrade *UpdateClusterRequestAutoUpgrade `json:"auto_upgrade"`
	// FeatureGates: list of feature gates to enable.
	FeatureGates *[]string `json:"feature_gates"`
	// AdmissionPlugins: list of admission plugins to enable.
	AdmissionPlugins *[]string `json:"admission_plugins"`
	// OpenIDConnectConfig: openID Connect configuration of the cluster. This configuration enables to update the OpenID Connect configuration of the Kubernetes API server.
	OpenIDConnectConfig *UpdateClusterRequestOpenIDConnectConfig `json:"open_id_connect_config"`
	// ApiserverCertSans: additional Subject Alternative Names for the Kubernetes API server certificate.
	ApiserverCertSans *[]string `json:"apiserver_cert_sans"`
}

// UpdateCluster: update a Cluster.
// Update information on a specific Kubernetes cluster. You can update details such as its name, description, tags and configuration. To upgrade a cluster, you will need to use the dedicated endpoint.
func (s *API) UpdateCluster(req *UpdateClusterRequest, opts ...scw.RequestOption) (*Cluster, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Cluster

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteClusterRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// ClusterID: ID of the cluster to delete.
	ClusterID string `json:"-"`
	// WithAdditionalResources: defines whether all volumes (including retain volume type), empty Private Networks and Load Balancers with a name starting with the cluster ID will also be deleted.
	WithAdditionalResources bool `json:"-"`
}

// DeleteCluster: delete a Cluster.
// Delete a specific Kubernetes cluster and all its associated pools and nodes. Note that this method will not delete any Load Balancer or Block Volume that are associated with the cluster.
func (s *API) DeleteCluster(req *DeleteClusterRequest, opts ...scw.RequestOption) (*Cluster, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	query := url.Values{}
	parameter.AddToQuery(query, "with_additional_resources", req.WithAdditionalResources)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "",
		Query:   query,
		Headers: http.Header{},
	}

	var resp Cluster

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpgradeClusterRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// ClusterID: ID of the cluster to upgrade.
	ClusterID string `json:"-"`
	// Version: new Kubernetes version of the cluster. Note that the version should either be a higher patch version of the same minor version or the direct minor version after the current one.
	Version string `json:"version"`
	// UpgradePools: defines whether pools will also be upgraded once the control plane is upgraded.
	UpgradePools bool `json:"upgrade_pools"`
}

// UpgradeCluster: upgrade a Cluster.
// Upgrade a specific Kubernetes cluster and possibly its associated pools to a specific and supported Kubernetes version.
func (s *API) UpgradeCluster(req *UpgradeClusterRequest, opts ...scw.RequestOption) (*Cluster, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "/upgrade",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Cluster

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type SetClusterTypeRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// ClusterID: ID of the cluster to migrate from one type to another.
	ClusterID string `json:"-"`
	// Type: type of the cluster. Note that some migrations are not possible (please refer to product documentation).
	Type string `json:"type"`
}

// SetClusterType: change the Cluster type.
// Change the type of a specific Kubernetes cluster. To see the possible values you can enter for the `type` field, [list available cluster types](#path-clusters-list-available-cluster-types-for-a-cluster).
func (s *API) SetClusterType(req *SetClusterTypeRequest, opts ...scw.RequestOption) (*Cluster, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "/set-type",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Cluster

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListClusterAvailableVersionsRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// ClusterID: cluster ID for which the available Kubernetes versions will be listed.
	ClusterID string `json:"-"`
}

// ListClusterAvailableVersions: list available versions for a Cluster.
// List the versions that a specific Kubernetes cluster is allowed to upgrade to. Results will include every patch version greater than the current patch, as well as one minor version ahead of the current version. Any upgrade skipping a minor version will not work.
func (s *API) ListClusterAvailableVersions(req *ListClusterAvailableVersionsRequest, opts ...scw.RequestOption) (*ListClusterAvailableVersionsResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "/available-versions",
		Headers: http.Header{},
	}

	var resp ListClusterAvailableVersionsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListClusterAvailableTypesRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// ClusterID: cluster ID for which the available Kubernetes types will be listed.
	ClusterID string `json:"-"`
}

// ListClusterAvailableTypes: list available cluster types for a cluster.
// List the cluster types that a specific Kubernetes cluster is allowed to switch to.
func (s *API) ListClusterAvailableTypes(req *ListClusterAvailableTypesRequest, opts ...scw.RequestOption) (*ListClusterAvailableTypesResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "/available-types",
		Headers: http.Header{},
	}

	var resp ListClusterAvailableTypesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetClusterKubeConfigRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// ClusterID: cluster ID for which to download the kubeconfig.
	ClusterID string `json:"-"`
}

// getClusterKubeConfig: download the kubeconfig for a Cluster.
// Download the Kubernetes cluster config file (also known as `kubeconfig`) for a specific cluster to use it with `kubectl`.
// Tip: add `?dl=1` at the end of the URL to directly retrieve the base64 decoded kubeconfig. If you choose not to, the kubeconfig will be base64 encoded.
func (s *API) getClusterKubeConfig(req *GetClusterKubeConfigRequest, opts ...scw.RequestOption) (*scw.File, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "/kubeconfig",
		Headers: http.Header{},
	}

	var resp scw.File

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ResetClusterAdminTokenRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// ClusterID: cluster ID on which the admin token will be renewed.
	ClusterID string `json:"-"`
}

// ResetClusterAdminToken: reset the admin token of a Cluster.
// Reset the admin token for a specific Kubernetes cluster. This will revoke the old admin token (which will not be usable afterwards) and create a new one. Note that you will need to download kubeconfig again to keep interacting with the cluster.
func (s *API) ResetClusterAdminToken(req *ResetClusterAdminTokenRequest, opts ...scw.RequestOption) error {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "/reset-admin-token",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return err
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type MigrateToPrivateNetworkClusterRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// ClusterID: ID of the cluster to migrate.
	ClusterID string `json:"-"`
	// PrivateNetworkID: ID of the Private Network to link to the cluster.
	PrivateNetworkID string `json:"private_network_id"`
}

// MigrateToPrivateNetworkCluster: migrate an existing cluster to a Private Network cluster.
// Migrate a cluster that was created before the release of Private Network clusters to a new one with a Private Network.
func (s *API) MigrateToPrivateNetworkCluster(req *MigrateToPrivateNetworkClusterRequest, opts ...scw.RequestOption) (*Cluster, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "/migrate-to-private-network",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Cluster

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListPoolsRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// ClusterID: ID of the cluster whose pools will be listed.
	ClusterID string `json:"-"`
	// OrderBy: sort order of returned pools.
	// Default value: created_at_asc
	OrderBy ListPoolsRequestOrderBy `json:"-"`
	// Page: page number for the returned pools.
	Page *int32 `json:"-"`
	// PageSize: maximum number of pools per page.
	PageSize *uint32 `json:"-"`
	// Name: name to filter on, only pools containing this substring in their name will be returned.
	Name *string `json:"-"`
	// Status: status to filter on, only pools with this status will be returned.
	// Default value: unknown
	Status PoolStatus `json:"-"`
}

// ListPools: list Pools in a Cluster.
// List all the existing pools for a specific Kubernetes cluster.
func (s *API) ListPools(req *ListPoolsRequest, opts ...scw.RequestOption) (*ListPoolsResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "status", req.Status)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "/pools",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListPoolsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreatePoolRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// ClusterID: cluster ID to which the pool will be attached.
	ClusterID string `json:"-"`
	// Name: pool name.
	Name string `json:"name"`
	// NodeType: node type is the type of Scaleway Instance wanted for the pool. Nodes with insufficient memory are not eligible (DEV1-S, PLAY2-PICO, STARDUST). 'external' is a special node type used to provision instances from other cloud providers in a Kosmos Cluster.
	NodeType string `json:"node_type"`
	// PlacementGroupID: placement group ID in which all the nodes of the pool will be created.
	PlacementGroupID *string `json:"placement_group_id"`
	// Autoscaling: defines whether the autoscaling feature is enabled for the pool.
	Autoscaling bool `json:"autoscaling"`
	// Size: size (number of nodes) of the pool.
	Size uint32 `json:"size"`
	// MinSize: defines the minimum size of the pool. Note that this field is only used when autoscaling is enabled on the pool.
	MinSize *uint32 `json:"min_size"`
	// MaxSize: defines the maximum size of the pool. Note that this field is only used when autoscaling is enabled on the pool.
	MaxSize *uint32 `json:"max_size"`
	// ContainerRuntime: customization of the container runtime is available for each pool. Note that `docker` has been deprecated since version 1.20 and will be removed by version 1.24.
	// Default value: unknown_runtime
	ContainerRuntime Runtime `json:"container_runtime"`
	// Autohealing: defines whether the autohealing feature is enabled for the pool.
	Autohealing bool `json:"autohealing"`
	// Tags: tags associated with the pool.
	Tags []string `json:"tags"`
	// KubeletArgs: kubelet arguments to be used by this pool. Note that this feature is experimental.
	KubeletArgs map[string]string `json:"kubelet_args"`
	// UpgradePolicy: pool upgrade policy.
	UpgradePolicy *CreatePoolRequestUpgradePolicy `json:"upgrade_policy"`
	// Zone: zone in which the pool's nodes will be spawned.
	Zone scw.Zone `json:"zone"`
	// RootVolumeType: defines the system volume disk type. Two different types of volume (`volume_type`) are provided: `l_ssd` is a local block storage which means your system is stored locally on your node's hypervisor. `b_ssd` is a remote block storage which means your system is stored on a centralized and resilient cluster.
	// Default value: default_volume_type
	RootVolumeType PoolVolumeType `json:"root_volume_type"`
	// RootVolumeSize: system volume disk size.
	RootVolumeSize *scw.Size `json:"root_volume_size"`
}

// CreatePool: create a new Pool in a Cluster.
// Create a new pool in a specific Kubernetes cluster.
func (s *API) CreatePool(req *CreatePoolRequest, opts ...scw.RequestOption) (*Pool, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if req.Zone == "" {
		defaultZone, _ := s.client.GetDefaultZone()
		req.Zone = defaultZone
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("pool")
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "/pools",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Pool

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetPoolRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// PoolID: ID of the requested pool.
	PoolID string `json:"-"`
}

// GetPool: get a Pool in a Cluster.
// Retrieve details about a specific pool in a Kubernetes cluster.
func (s *API) GetPool(req *GetPoolRequest, opts ...scw.RequestOption) (*Pool, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.PoolID) == "" {
		return nil, errors.New("field PoolID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/pools/" + fmt.Sprint(req.PoolID) + "",
		Headers: http.Header{},
	}

	var resp Pool

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpgradePoolRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// PoolID: ID of the pool to upgrade.
	PoolID string `json:"-"`
	// Version: new Kubernetes version for the pool.
	Version string `json:"version"`
}

// UpgradePool: upgrade a Pool in a Cluster.
// Upgrade the Kubernetes version of a specific pool. Note that it only works if the targeted version matches the cluster's version.
func (s *API) UpgradePool(req *UpgradePoolRequest, opts ...scw.RequestOption) (*Pool, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.PoolID) == "" {
		return nil, errors.New("field PoolID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/pools/" + fmt.Sprint(req.PoolID) + "/upgrade",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Pool

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdatePoolRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// PoolID: ID of the pool to update.
	PoolID string `json:"-"`
	// Autoscaling: new value for the pool autoscaling enablement.
	Autoscaling *bool `json:"autoscaling"`
	// Size: new desired pool size.
	Size *uint32 `json:"size"`
	// MinSize: new minimum size for the pool.
	MinSize *uint32 `json:"min_size"`
	// MaxSize: new maximum size for the pool.
	MaxSize *uint32 `json:"max_size"`
	// Autohealing: new value for the pool autohealing enablement.
	Autohealing *bool `json:"autohealing"`
	// Tags: new tags associated with the pool.
	Tags *[]string `json:"tags"`
	// KubeletArgs: new Kubelet arguments to be used by this pool. Note that this feature is experimental.
	KubeletArgs *map[string]string `json:"kubelet_args"`
	// UpgradePolicy: new upgrade policy for the pool.
	UpgradePolicy *UpdatePoolRequestUpgradePolicy `json:"upgrade_policy"`
}

// UpdatePool: update a Pool in a Cluster.
// Update the attributes of a specific pool, such as its desired size, autoscaling settings, and tags.
func (s *API) UpdatePool(req *UpdatePoolRequest, opts ...scw.RequestOption) (*Pool, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.PoolID) == "" {
		return nil, errors.New("field PoolID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/pools/" + fmt.Sprint(req.PoolID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Pool

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeletePoolRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// PoolID: ID of the pool to delete.
	PoolID string `json:"-"`
}

// DeletePool: delete a Pool in a Cluster.
// Delete a specific pool from a cluster. Note that all the pool's nodes will also be deleted.
func (s *API) DeletePool(req *DeletePoolRequest, opts ...scw.RequestOption) (*Pool, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.PoolID) == "" {
		return nil, errors.New("field PoolID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/pools/" + fmt.Sprint(req.PoolID) + "",
		Headers: http.Header{},
	}

	var resp Pool

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type CreateExternalNodeRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`

	PoolID string `json:"-"`
}

// CreateExternalNode: create a Kosmos node.
// Retrieve metadata for a Kosmos node. This method is not intended to be called by end users but rather programmatically by the kapsule-node-agent.
func (s *API) CreateExternalNode(req *CreateExternalNodeRequest, opts ...scw.RequestOption) (*ExternalNode, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.PoolID) == "" {
		return nil, errors.New("field PoolID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/pools/" + fmt.Sprint(req.PoolID) + "/external-nodes",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp ExternalNode

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListNodesRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// ClusterID: cluster ID from which the nodes will be listed from.
	ClusterID string `json:"-"`
	// PoolID: pool ID on which to filter the returned nodes.
	PoolID *string `json:"-"`
	// OrderBy: sort order of the returned nodes.
	// Default value: created_at_asc
	OrderBy ListNodesRequestOrderBy `json:"-"`
	// Page: page number for the returned nodes.
	Page *int32 `json:"-"`
	// PageSize: maximum number of nodes per page.
	PageSize *uint32 `json:"-"`
	// Name: name to filter on, only nodes containing this substring in their name will be returned.
	Name *string `json:"-"`
	// Status: status to filter on, only nodes with this status will be returned.
	// Default value: unknown
	Status NodeStatus `json:"-"`
}

// ListNodes: list Nodes in a Cluster.
// List all the existing nodes for a specific Kubernetes cluster.
func (s *API) ListNodes(req *ListNodesRequest, opts ...scw.RequestOption) (*ListNodesResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "pool_id", req.PoolID)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "status", req.Status)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.ClusterID) == "" {
		return nil, errors.New("field ClusterID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/clusters/" + fmt.Sprint(req.ClusterID) + "/nodes",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListNodesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetNodeRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// NodeID: ID of the requested node.
	NodeID string `json:"-"`
}

// GetNode: get a Node in a Cluster.
// Retrieve details about a specific Kubernetes Node.
func (s *API) GetNode(req *GetNodeRequest, opts ...scw.RequestOption) (*Node, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.NodeID) == "" {
		return nil, errors.New("field NodeID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/nodes/" + fmt.Sprint(req.NodeID) + "",
		Headers: http.Header{},
	}

	var resp Node

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ReplaceNodeRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// NodeID: ID of the node to replace.
	NodeID string `json:"-"`
}

// Deprecated: ReplaceNode: replace a Node in a Cluster.
// Replace a specific Node. The node will first be cordoned (scheduling will be disabled on it). The existing pods on the node will then be drained and rescheduled onto another schedulable node. Note that when there is not enough space to reschedule all the pods (such as in a one-node cluster), disruption of your applications can be expected.
func (s *API) ReplaceNode(req *ReplaceNodeRequest, opts ...scw.RequestOption) (*Node, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.NodeID) == "" {
		return nil, errors.New("field NodeID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/nodes/" + fmt.Sprint(req.NodeID) + "/replace",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Node

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type RebootNodeRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// NodeID: ID of the node to reboot.
	NodeID string `json:"-"`
}

// RebootNode: reboot a Node in a Cluster.
// Reboot a specific Node. The node will first be cordoned (scheduling will be disabled on it). The existing pods on the node will then be drained and rescheduled onto another schedulable node. Note that when there is not enough space to reschedule all the pods (such as in a one-node cluster), disruption of your applications can be expected.
func (s *API) RebootNode(req *RebootNodeRequest, opts ...scw.RequestOption) (*Node, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.NodeID) == "" {
		return nil, errors.New("field NodeID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/nodes/" + fmt.Sprint(req.NodeID) + "/reboot",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Node

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type DeleteNodeRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// NodeID: ID of the node to replace.
	NodeID string `json:"-"`
	// SkipDrain: skip draining node from its workload.
	SkipDrain bool `json:"-"`
	// Replace: add a new node after the deletion of this node.
	Replace bool `json:"-"`
}

// DeleteNode: delete a Node in a Cluster.
// Delete a specific Node. Note that when there is not enough space to reschedule all the pods (such as in a one-node cluster), disruption of your applications can be expected.
func (s *API) DeleteNode(req *DeleteNodeRequest, opts ...scw.RequestOption) (*Node, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	query := url.Values{}
	parameter.AddToQuery(query, "skip_drain", req.SkipDrain)
	parameter.AddToQuery(query, "replace", req.Replace)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.NodeID) == "" {
		return nil, errors.New("field NodeID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/nodes/" + fmt.Sprint(req.NodeID) + "",
		Query:   query,
		Headers: http.Header{},
	}

	var resp Node

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListVersionsRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
}

// ListVersions: list all available Versions.
// List all available versions for the creation of a new Kubernetes cluster.
func (s *API) ListVersions(req *ListVersionsRequest, opts ...scw.RequestOption) (*ListVersionsResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/versions",
		Headers: http.Header{},
	}

	var resp ListVersionsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetVersionRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// VersionName: requested version name.
	VersionName string `json:"-"`
}

// GetVersion: get a Version.
// Retrieve a specific Kubernetes version and its details.
func (s *API) GetVersion(req *GetVersionRequest, opts ...scw.RequestOption) (*Version, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	if fmt.Sprint(req.VersionName) == "" {
		return nil, errors.New("field VersionName cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/versions/" + fmt.Sprint(req.VersionName) + "",
		Headers: http.Header{},
	}

	var resp Version

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ListClusterTypesRequest struct {
	// Region: region to target. If none is passed will use default region from the config.
	Region scw.Region `json:"-"`
	// Page: page number, from the paginated results, to return for cluster-types.
	Page *int32 `json:"-"`
	// PageSize: maximum number of clusters per page.
	PageSize *uint32 `json:"-"`
}

// ListClusterTypes: list cluster types.
// List available cluster types and their technical details.
func (s *API) ListClusterTypes(req *ListClusterTypesRequest, opts ...scw.RequestOption) (*ListClusterTypesResponse, error) {
	var err error

	if req.Region == "" {
		defaultRegion, _ := s.client.GetDefaultRegion()
		req.Region = defaultRegion
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)

	if fmt.Sprint(req.Region) == "" {
		return nil, errors.New("field Region cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/k8s/v1/regions/" + fmt.Sprint(req.Region) + "/cluster-types",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListClusterTypesResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListClustersResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListClustersResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListClustersResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Clusters = append(r.Clusters, results.Clusters...)
	r.TotalCount += uint32(len(results.Clusters))
	return uint32(len(results.Clusters)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListPoolsResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListPoolsResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListPoolsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Pools = append(r.Pools, results.Pools...)
	r.TotalCount += uint32(len(results.Pools))
	return uint32(len(results.Pools)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListNodesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListNodesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListNodesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Nodes = append(r.Nodes, results.Nodes...)
	r.TotalCount += uint32(len(results.Nodes))
	return uint32(len(results.Nodes)), nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListClusterTypesResponse) UnsafeGetTotalCount() uint32 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListClusterTypesResponse) UnsafeAppend(res interface{}) (uint32, error) {
	results, ok := res.(*ListClusterTypesResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.ClusterTypes = append(r.ClusterTypes, results.ClusterTypes...)
	r.TotalCount += uint32(len(results.ClusterTypes))
	return uint32(len(results.ClusterTypes)), nil
}
