package model

import "time"

// AccessList represents a reusable access control list
type AccessList struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	SatisfyAny  bool             `json:"satisfy_any"` // true = any rule match, false = all must match
	PassAuth    bool             `json:"pass_auth"`   // Allow authenticated users to bypass
	Items       []AccessListItem `json:"items,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// AccessListItem represents a single allow/deny rule
type AccessListItem struct {
	ID           string    `json:"id"`
	AccessListID string    `json:"access_list_id"`
	Directive    string    `json:"directive"` // "allow" or "deny"
	Address      string    `json:"address"`   // IP, CIDR, or "all"
	Description  string    `json:"description,omitempty"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateAccessListRequest is the request to create an access list
type CreateAccessListRequest struct {
	Name        string                       `json:"name" validate:"required,min=1,max=255"`
	Description string                       `json:"description,omitempty"`
	SatisfyAny  *bool                        `json:"satisfy_any,omitempty"`
	PassAuth    *bool                        `json:"pass_auth,omitempty"`
	Items       []CreateAccessListItemRequest `json:"items,omitempty"`
}

// CreateAccessListItemRequest is the request to create an access list item
type CreateAccessListItemRequest struct {
	Directive   string `json:"directive" validate:"required,oneof=allow deny"`
	Address     string `json:"address" validate:"required"`
	Description string `json:"description,omitempty"`
	SortOrder   int    `json:"sort_order,omitempty"`
}

// UpdateAccessListRequest is the request to update an access list
type UpdateAccessListRequest struct {
	Name        *string                       `json:"name,omitempty"`
	Description *string                       `json:"description,omitempty"`
	SatisfyAny  *bool                         `json:"satisfy_any,omitempty"`
	PassAuth    *bool                         `json:"pass_auth,omitempty"`
	Items       []CreateAccessListItemRequest `json:"items,omitempty"`
}

// AccessListListResponse is the response for listing access lists
type AccessListListResponse struct {
	Data       []AccessList `json:"data"`
	Total      int          `json:"total"`
	Page       int          `json:"page"`
	PerPage    int          `json:"per_page"`
	TotalPages int          `json:"total_pages"`
}
