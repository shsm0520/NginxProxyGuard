package model

// ProxyHostListFilter narrows GET /proxy-hosts. Every field is optional; Tags
// is AND semantics (the host must carry all of them). Values arrive already
// normalised/validated by the handler.
type ProxyHostListFilter struct {
	Tags     []string
	Domain   string // parent domain of the first domain name, e.g. "example.com"
	Upstream string // exact forward_host
	Enabled  *bool
}

// ProxyHostGroupCount is one bucket of an auto-group.
type ProxyHostGroupCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type ProxyHostGroupStatus struct {
	Enabled  int `json:"enabled"`
	Disabled int `json:"disabled"`
}

// ProxyHostGroups is the read-only summary behind the list's filter panel.
// Arrays are sorted by count desc, then name, and are never nil so the JSON
// always carries [] rather than null.
type ProxyHostGroups struct {
	Tags      []ProxyHostGroupCount `json:"tags"`
	Domains   []ProxyHostGroupCount `json:"domains"`
	Upstreams []ProxyHostGroupCount `json:"upstreams"`
	Status    ProxyHostGroupStatus  `json:"status"`
}
