package api

import (
	"context"
	"fmt"
	"net/url"
)

// ─── /me ────────────────────────────────────────────────────────────────────

// MeResponse represents the current user's full profile
type MeResponse struct {
	ID       string   `json:"id"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	TenantID string   `json:"tenant_id"`
	Roles    []string `json:"roles"`
	Groups   []string `json:"groups"`
	Teams    []string `json:"teams"`
}

// GetMe fetches the current user's profile
func (c *Client) GetMe(ctx context.Context) (*MeResponse, error) {
	var resp MeResponse
	if err := c.Get(ctx, "/api/v1/me", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ─── Entities ────────────────────────────────────────────────────────────────

// Entity represents a catalog entity (summary form)
type Entity struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Type        string   `json:"type"`
	Owner       string   `json:"owner"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	TenantID    string   `json:"tenant_id"`
}

// EntityDetail represents full entity detail with all sub-resources
type EntityDetail struct {
	Entity
	Links     []EntityLink `json:"links"`
	Lifecycle string       `json:"lifecycle"`
	Tier      string       `json:"tier"`
}

// EntityLink represents a link on an entity
type EntityLink struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Icon  string `json:"icon"`
}

// Resource represents an entity resource (DB, queue, etc.)
type Resource struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Environment string `json:"environment"`
	Description string `json:"description"`
}

// EntityStatus represents entity health/status
type EntityStatus struct {
	Health        string  `json:"health"`
	Uptime        float64 `json:"uptime"`
	LastDeployAt  string  `json:"last_deploy_at"`
	IncidentCount int     `json:"incident_count"`
}

// ChangelogEntry represents a single changelog item
type ChangelogEntry struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Timestamp   string `json:"timestamp"`
}

// ListEntitiesOpts holds optional filters for listing entities
type ListEntitiesOpts struct {
	Type   string
	Search string
	Owner  string
}

// EntitiesResponse is the paginated response from /entities
type EntitiesResponse struct {
	Entities   []Entity `json:"entities"`
	TotalCount int      `json:"total_count"`
	NextCursor string   `json:"next_cursor"`
}

// ListEntities returns all entities matching the given filters (handles pagination)
func (c *Client) ListEntities(ctx context.Context, opts ListEntitiesOpts) ([]*Entity, error) {
	q := url.Values{}
	if opts.Type != "" {
		q.Set("type", opts.Type)
	}
	if opts.Search != "" {
		q.Set("search", opts.Search)
	}
	if opts.Owner != "" {
		q.Set("owner", opts.Owner)
	}
	q.Set("limit", "100")

	path := "/api/v1/entities?" + q.Encode()

	var resp EntitiesResponse
	if err := c.Get(ctx, path, &resp); err != nil {
		return nil, err
	}

	entities := make([]*Entity, len(resp.Entities))
	for i := range resp.Entities {
		e := resp.Entities[i]
		entities[i] = &e
	}
	return entities, nil
}

// GetEntity fetches a single entity by ID or slug
func (c *Client) GetEntity(ctx context.Context, id string) (*EntityDetail, error) {
	var resp EntityDetail
	if err := c.Get(ctx, "/api/v1/entities/"+id, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetEntityResources fetches an entity's associated resources
func (c *Client) GetEntityResources(ctx context.Context, id string) ([]*Resource, error) {
	var resp struct {
		Resources []Resource `json:"resources"`
	}
	if err := c.Get(ctx, fmt.Sprintf("/api/v1/entities/%s/resources", id), &resp); err != nil {
		return nil, err
	}
	resources := make([]*Resource, len(resp.Resources))
	for i := range resp.Resources {
		r := resp.Resources[i]
		resources[i] = &r
	}
	return resources, nil
}

// GetEntityStatus fetches an entity's live health/status
func (c *Client) GetEntityStatus(ctx context.Context, id string) (*EntityStatus, error) {
	var resp EntityStatus
	if err := c.Get(ctx, fmt.Sprintf("/api/v1/entities/%s/status", id), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetEntityChangelog fetches an entity's changelog entries
func (c *Client) GetEntityChangelog(ctx context.Context, id string) ([]*ChangelogEntry, error) {
	var resp struct {
		Entries []ChangelogEntry `json:"entries"`
	}
	if err := c.Get(ctx, fmt.Sprintf("/api/v1/entities/%s/changelog", id), &resp); err != nil {
		return nil, err
	}
	entries := make([]*ChangelogEntry, len(resp.Entries))
	for i := range resp.Entries {
		e := resp.Entries[i]
		entries[i] = &e
	}
	return entries, nil
}

// ─── Scorecard ───────────────────────────────────────────────────────────────

// Scorecard represents an entity's scorecard result
type Scorecard struct {
	Score      int             `json:"score"`
	Grade      string          `json:"grade"`
	MaxScore   int             `json:"max_score"`
	Checks     []ScorecardCheck `json:"checks"`
	UpdatedAt  string          `json:"updated_at"`
}

// ScorecardCheck is a single check in a scorecard
type ScorecardCheck struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Weight  int    `json:"weight"`
	Message string `json:"message"`
}

// GetEntityScorecard fetches an entity's scorecard
func (c *Client) GetEntityScorecard(ctx context.Context, id string) (*Scorecard, error) {
	var resp Scorecard
	if err := c.Get(ctx, fmt.Sprintf("/api/v1/entities/%s/scorecard", id), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ─── Teams ───────────────────────────────────────────────────────────────────

// Team represents a team summary
type Team struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	MemberCount int    `json:"member_count"`
}

// TeamDetail includes members and other team details
type TeamDetail struct {
	Team
	Members []TeamMember `json:"members"`
}

// TeamMember represents a member in a team
type TeamMember struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

// TeamsResponse is the response from /teams
type TeamsResponse struct {
	Teams []Team `json:"teams"`
}

// ListTeams returns all teams
func (c *Client) ListTeams(ctx context.Context) ([]*Team, error) {
	var resp TeamsResponse
	if err := c.Get(ctx, "/api/v1/teams", &resp); err != nil {
		return nil, err
	}
	teams := make([]*Team, len(resp.Teams))
	for i := range resp.Teams {
		t := resp.Teams[i]
		teams[i] = &t
	}
	return teams, nil
}

// GetTeam fetches a team by ID or slug (including members)
func (c *Client) GetTeam(ctx context.Context, idOrSlug string) (*TeamDetail, error) {
	var wrapper struct {
		Team    Team         `json:"team"`
		Members []TeamMember `json:"members"`
	}
	if err := c.Get(ctx, "/api/v1/teams/"+idOrSlug, &wrapper); err != nil {
		return nil, err
	}
	return &TeamDetail{
		Team:    wrapper.Team,
		Members: wrapper.Members,
	}, nil
}

// ─── Users ───────────────────────────────────────────────────────────────────

// User represents a user summary
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// UserDetail includes groups, teams, and roles
type UserDetail struct {
	User
	Groups []string `json:"groups"`
	Teams  []string `json:"teams"`
	Roles  []string `json:"roles"`
}

// UsersResponse is the response from /users
type UsersResponse struct {
	Users []User `json:"users"`
}

// ListUsers returns all users in the directory
func (c *Client) ListUsers(ctx context.Context) ([]*User, error) {
	var resp UsersResponse
	if err := c.Get(ctx, "/api/v1/users", &resp); err != nil {
		return nil, err
	}
	users := make([]*User, len(resp.Users))
	for i := range resp.Users {
		u := resp.Users[i]
		users[i] = &u
	}
	return users, nil
}

// GetUser fetches a single user by ID
func (c *Client) GetUser(ctx context.Context, id string) (*UserDetail, error) {
	var resp UserDetail
	if err := c.Get(ctx, "/api/v1/users/"+id, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ─── Groups ──────────────────────────────────────────────────────────────────

// Group represents a directory group
type Group struct {
	Name      string `json:"name"`
	RoleCount int    `json:"role_count"`
}

// Role represents a platform role
type Role struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GroupsResponse is the response from /groups
type GroupsResponse struct {
	Groups []Group `json:"groups"`
}

// RolesResponse is the response from /groups/{name}/roles
type RolesResponse struct {
	Roles []Role `json:"roles"`
}

// ListGroups returns all groups
func (c *Client) ListGroups(ctx context.Context) ([]*Group, error) {
	var resp GroupsResponse
	if err := c.Get(ctx, "/api/v1/groups", &resp); err != nil {
		return nil, err
	}
	groups := make([]*Group, len(resp.Groups))
	for i := range resp.Groups {
		g := resp.Groups[i]
		groups[i] = &g
	}
	return groups, nil
}

// GetGroupRoles fetches the roles mapped to a group
func (c *Client) GetGroupRoles(ctx context.Context, groupName string) ([]*Role, error) {
	var resp RolesResponse
	if err := c.Get(ctx, fmt.Sprintf("/api/v1/groups/%s/roles", groupName), &resp); err != nil {
		return nil, err
	}
	roles := make([]*Role, len(resp.Roles))
	for i := range resp.Roles {
		r := resp.Roles[i]
		roles[i] = &r
	}
	return roles, nil
}

// ─── Search ──────────────────────────────────────────────────────────────────

// SearchHit is a single search result item
type SearchHit struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Owner       string `json:"owner"`
	Score       float64 `json:"score"`
}

// SearchResult wraps search hits
type SearchResult struct {
	Hits       []SearchHit `json:"hits"`
	TotalCount int         `json:"total_count"`
}

// Search performs a full-text search across entities
func (c *Client) Search(ctx context.Context, query string) (*SearchResult, error) {
	q := url.Values{}
	q.Set("q", query)
	var resp SearchResult
	if err := c.Get(ctx, "/api/v1/search?"+q.Encode(), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ─── K8s ─────────────────────────────────────────────────────────────────────

// K8sAgent represents a connected K8s agent
type K8sAgent struct {
	ID          string `json:"id"`
	ClusterName string `json:"cluster_name"`
	Status      string `json:"status"`
	Version     string `json:"version"`
	LastSeen    string `json:"last_seen"`
	Namespace   string `json:"namespace"`
}

// K8sAgentsResponse is the response from /k8s/agents
type K8sAgentsResponse struct {
	Agents []K8sAgent `json:"agents"`
}

// ListK8sAgents returns all registered K8s agents
func (c *Client) ListK8sAgents(ctx context.Context) ([]*K8sAgent, error) {
	var resp K8sAgentsResponse
	if err := c.Get(ctx, "/api/v1/k8s/agents", &resp); err != nil {
		return nil, err
	}
	agents := make([]*K8sAgent, len(resp.Agents))
	for i := range resp.Agents {
		a := resp.Agents[i]
		agents[i] = &a
	}
	return agents, nil
}

// ─── Forge ───────────────────────────────────────────────────────────────────

// Mold represents a Forge mold (workflow template)
type Mold struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

// MoldInput describes a single input parameter for a mold
type MoldInput struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     string `json:"default"`
}

// MoldStep describes a single step in a mold
type MoldStep struct {
	Name   string `json:"name"`
	Action string `json:"action"`
}

// MoldDetail is the full mold definition
type MoldDetail struct {
	Mold
	Inputs []MoldInput `json:"inputs"`
	Steps  []MoldStep  `json:"steps"`
}

// MoldsResponse is the response from /forge/molds
type MoldsResponse struct {
	Molds []Mold `json:"molds"`
}

// ForgeRun represents a workflow run (canonical type for the api package)
type ForgeRun struct {
	ID          string `json:"id"`
	MoldID      string `json:"mold_id"`
	MoldSlug    string `json:"mold_slug"`
	Status      string `json:"status"`
	CreatedBy   string `json:"created_by"`
	CreatedAt   string `json:"created_at"`
	CompletedAt string `json:"completed_at,omitempty"`
	Error       string `json:"error,omitempty"`
}

// ForgeRunsResponse is the response from /forge/runs
type ForgeRunsResponse struct {
	Runs       []ForgeRun `json:"runs"`
	TotalCount int        `json:"total_count"`
}

// CreateRunRequest is the body for POST /forge/runs
type CreateRunRequest struct {
	MoldSlug string         `json:"mold_slug"`
	Inputs   map[string]any `json:"inputs,omitempty"`
}

// ListMolds returns all forge molds
func (c *Client) ListMolds(ctx context.Context) ([]*Mold, error) {
	var resp MoldsResponse
	if err := c.Get(ctx, "/api/v1/forge/molds", &resp); err != nil {
		return nil, err
	}
	molds := make([]*Mold, len(resp.Molds))
	for i := range resp.Molds {
		m := resp.Molds[i]
		molds[i] = &m
	}
	return molds, nil
}

// GetMold fetches a single mold by slug
func (c *Client) GetMold(ctx context.Context, slug string) (*MoldDetail, error) {
	var resp MoldDetail
	if err := c.Get(ctx, "/api/v1/forge/molds/"+slug, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateRun starts a new forge run from a mold slug
func (c *Client) CreateRun(ctx context.Context, moldSlug string, inputs map[string]any) (*ForgeRun, error) {
	req := CreateRunRequest{MoldSlug: moldSlug, Inputs: inputs}
	var resp ForgeRun
	if err := c.Post(ctx, "/api/v1/forge/runs", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
