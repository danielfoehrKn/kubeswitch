// This file was automatically generated. DO NOT EDIT.
// If you have any remark or suggestion do not hesitate to open an issue.

// Package account provides methods and message types of the account v3 API.
package account

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

// ProjectAPI: this API allows you to manage projects.
type ProjectAPI struct {
	client *scw.Client
}

// NewProjectAPI returns a ProjectAPI object from a Scaleway client.
func NewProjectAPI(client *scw.Client) *ProjectAPI {
	return &ProjectAPI{
		client: client,
	}
}

type ListProjectsRequestOrderBy string

const (
	// Creation date ascending
	ListProjectsRequestOrderByCreatedAtAsc = ListProjectsRequestOrderBy("created_at_asc")
	// Creation date descending
	ListProjectsRequestOrderByCreatedAtDesc = ListProjectsRequestOrderBy("created_at_desc")
	// Name ascending
	ListProjectsRequestOrderByNameAsc = ListProjectsRequestOrderBy("name_asc")
	// Name descending
	ListProjectsRequestOrderByNameDesc = ListProjectsRequestOrderBy("name_desc")
)

func (enum ListProjectsRequestOrderBy) String() string {
	if enum == "" {
		// return default value if empty
		return "created_at_asc"
	}
	return string(enum)
}

func (enum ListProjectsRequestOrderBy) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum)), nil
}

func (enum *ListProjectsRequestOrderBy) UnmarshalJSON(data []byte) error {
	tmp := ""

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*enum = ListProjectsRequestOrderBy(ListProjectsRequestOrderBy(tmp).String())
	return nil
}

// ListProjectsResponse: list projects response.
type ListProjectsResponse struct {
	// TotalCount: total number of Projects.
	TotalCount uint64 `json:"total_count"`
	// Projects: paginated returned Projects.
	Projects []*Project `json:"projects"`
}

// Project: project.
type Project struct {
	// ID: ID of the Project.
	ID string `json:"id"`
	// Name: name of the Project.
	Name string `json:"name"`
	// OrganizationID: organization ID of the Project.
	OrganizationID string `json:"organization_id"`
	// CreatedAt: creation date of the Project.
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt: update date of the Project.
	UpdatedAt *time.Time `json:"updated_at"`
	// Description: description of the Project.
	Description string `json:"description"`
}

// Service ProjectAPI

type ProjectAPICreateProjectRequest struct {
	// Name: name of the Project.
	Name string `json:"name"`
	// OrganizationID: organization ID of the Project.
	OrganizationID string `json:"organization_id"`
	// Description: description of the Project.
	Description string `json:"description"`
}

// CreateProject: create a new Project for an Organization.
// Generate a new Project for an Organization, specifying its configuration including name and description.
func (s *ProjectAPI) CreateProject(req *ProjectAPICreateProjectRequest, opts ...scw.RequestOption) (*Project, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	if req.Name == "" {
		req.Name = namegenerator.GetRandomName("proj")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "POST",
		Path:    "/account/v3/projects",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Project

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ProjectAPIListProjectsRequest struct {
	// OrganizationID: organization ID of the Project.
	OrganizationID string `json:"-"`
	// Name: name of the Project.
	Name *string `json:"-"`
	// Page: page number for the returned Projects.
	Page *int32 `json:"-"`
	// PageSize: maximum number of Project per page.
	PageSize *uint32 `json:"-"`
	// OrderBy: sort order of the returned Projects.
	// Default value: created_at_asc
	OrderBy ListProjectsRequestOrderBy `json:"-"`
	// ProjectIDs: project IDs to filter for. The results will be limited to any Projects with an ID in this array.
	ProjectIDs []string `json:"-"`
}

// ListProjects: list all Projects of an Organization.
// List all Projects of an Organization. The response will include the total number of Projects as well as their associated Organizations, names, and IDs. Other information includes the creation and update date of the Project.
func (s *ProjectAPI) ListProjects(req *ProjectAPIListProjectsRequest, opts ...scw.RequestOption) (*ListProjectsResponse, error) {
	var err error

	if req.OrganizationID == "" {
		defaultOrganizationID, _ := s.client.GetDefaultOrganizationID()
		req.OrganizationID = defaultOrganizationID
	}

	defaultPageSize, exist := s.client.GetDefaultPageSize()
	if (req.PageSize == nil || *req.PageSize == 0) && exist {
		req.PageSize = &defaultPageSize
	}

	query := url.Values{}
	parameter.AddToQuery(query, "organization_id", req.OrganizationID)
	parameter.AddToQuery(query, "name", req.Name)
	parameter.AddToQuery(query, "page", req.Page)
	parameter.AddToQuery(query, "page_size", req.PageSize)
	parameter.AddToQuery(query, "order_by", req.OrderBy)
	parameter.AddToQuery(query, "project_ids", req.ProjectIDs)

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/account/v3/projects",
		Query:   query,
		Headers: http.Header{},
	}

	var resp ListProjectsResponse

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ProjectAPIGetProjectRequest struct {
	// ProjectID: project ID of the Project.
	ProjectID string `json:"-"`
}

// GetProject: get an existing Project.
// Retrieve information about an existing Project, specified by its Project ID. Its full details, including ID, name and description, are returned in the response object.
func (s *ProjectAPI) GetProject(req *ProjectAPIGetProjectRequest, opts ...scw.RequestOption) (*Project, error) {
	var err error

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	if fmt.Sprint(req.ProjectID) == "" {
		return nil, errors.New("field ProjectID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "GET",
		Path:    "/account/v3/projects/" + fmt.Sprint(req.ProjectID) + "",
		Headers: http.Header{},
	}

	var resp Project

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type ProjectAPIDeleteProjectRequest struct {
	// ProjectID: project ID of the Project.
	ProjectID string `json:"-"`
}

// DeleteProject: delete an existing Project.
// Delete an existing Project, specified by its Project ID. The Project needs to be empty (meaning there are no resources left in it) to be deleted effectively. Note that deleting a Project is permanent, and cannot be undone.
func (s *ProjectAPI) DeleteProject(req *ProjectAPIDeleteProjectRequest, opts ...scw.RequestOption) error {
	var err error

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	if fmt.Sprint(req.ProjectID) == "" {
		return errors.New("field ProjectID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "DELETE",
		Path:    "/account/v3/projects/" + fmt.Sprint(req.ProjectID) + "",
		Headers: http.Header{},
	}

	err = s.client.Do(scwReq, nil, opts...)
	if err != nil {
		return err
	}
	return nil
}

type ProjectAPIUpdateProjectRequest struct {
	// ProjectID: project ID of the Project.
	ProjectID string `json:"-"`
	// Name: name of the Project.
	Name *string `json:"name"`
	// Description: description of the Project.
	Description *string `json:"description"`
}

// UpdateProject: update Project.
// Update the parameters of an existing Project, specified by its Project ID. These parameters include the name and description.
func (s *ProjectAPI) UpdateProject(req *ProjectAPIUpdateProjectRequest, opts ...scw.RequestOption) (*Project, error) {
	var err error

	if req.ProjectID == "" {
		defaultProjectID, _ := s.client.GetDefaultProjectID()
		req.ProjectID = defaultProjectID
	}

	if fmt.Sprint(req.ProjectID) == "" {
		return nil, errors.New("field ProjectID cannot be empty in request")
	}

	scwReq := &scw.ScalewayRequest{
		Method:  "PATCH",
		Path:    "/account/v3/projects/" + fmt.Sprint(req.ProjectID) + "",
		Headers: http.Header{},
	}

	err = scwReq.SetBody(req)
	if err != nil {
		return nil, err
	}

	var resp Project

	err = s.client.Do(scwReq, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnsafeGetTotalCount should not be used
// Internal usage only
func (r *ListProjectsResponse) UnsafeGetTotalCount() uint64 {
	return r.TotalCount
}

// UnsafeAppend should not be used
// Internal usage only
func (r *ListProjectsResponse) UnsafeAppend(res interface{}) (uint64, error) {
	results, ok := res.(*ListProjectsResponse)
	if !ok {
		return 0, errors.New("%T type cannot be appended to type %T", res, r)
	}

	r.Projects = append(r.Projects, results.Projects...)
	r.TotalCount += uint64(len(results.Projects))
	return uint64(len(results.Projects)), nil
}
