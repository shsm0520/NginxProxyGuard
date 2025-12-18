package handler

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type WAFTestHandler struct{}

func NewWAFTestHandler() *WAFTestHandler {
	return &WAFTestHandler{}
}

type WAFTestRequest struct {
	TargetURL   string `json:"target_url"`
	AttackType  string `json:"attack_type"`
	CustomValue string `json:"custom_value,omitempty"`
	HostHeader  string `json:"host_header,omitempty"` // Custom Host header for Docker internal testing
}

type WAFTestResult struct {
	AttackType  string `json:"attack_type"`
	TestURL     string `json:"test_url"`
	StatusCode  int    `json:"status_code"`
	Blocked     bool   `json:"blocked"`
	ResponseTime int64  `json:"response_time_ms"`
	Description string `json:"description"`
}

var attackPatterns = map[string]struct {
	Path        string
	Query       string
	Header      string
	HeaderValue string
	Description string
}{
	"sql_injection": {
		Query:       "id=1' OR '1'='1",
		Description: "SQL Injection - Classic OR-based bypass",
	},
	"sql_injection_union": {
		Query:       "id=1 UNION SELECT * FROM users--",
		Description: "SQL Injection - UNION-based attack",
	},
	"xss_script": {
		Query:       "q=<script>alert('XSS')</script>",
		Description: "XSS - Script tag injection",
	},
	"xss_event": {
		Query:       "q=<img src=x onerror=alert(1)>",
		Description: "XSS - Event handler injection",
	},
	"path_traversal": {
		Path:        "/../../etc/passwd",
		Description: "Path Traversal - Directory traversal attack",
	},
	"path_traversal_encoded": {
		Path:        "/%2e%2e/%2e%2e/etc/passwd",
		Description: "Path Traversal - URL encoded",
	},
	"command_injection": {
		Query:       "cmd=;cat /etc/passwd",
		Description: "Command Injection - Shell command",
	},
	"command_injection_pipe": {
		Query:       "cmd=|ls -la",
		Description: "Command Injection - Pipe command",
	},
	"scanner_sqlmap": {
		Header:      "User-Agent",
		HeaderValue: "sqlmap/1.0-dev",
		Description: "Scanner Detection - SQLMap user agent",
	},
	"scanner_nikto": {
		Header:      "User-Agent",
		HeaderValue: "Nikto/2.1.6",
		Description: "Scanner Detection - Nikto user agent",
	},
	"rce_php": {
		Query:       "file=php://filter/convert.base64-encode/resource=index.php",
		Description: "RCE - PHP wrapper attack",
	},
	"protocol_attack": {
		Header:      "X-Forwarded-Host",
		HeaderValue: "evil.com",
		Description: "Protocol Attack - Host header injection",
	},
}

func (h *WAFTestHandler) Test(w http.ResponseWriter, r *http.Request) {
	var req WAFTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TargetURL == "" {
		http.Error(w, "target_url is required", http.StatusBadRequest)
		return
	}

	pattern, exists := attackPatterns[req.AttackType]
	if !exists {
		http.Error(w, "Invalid attack_type", http.StatusBadRequest)
		return
	}

	// Build test URL
	testURL := req.TargetURL
	if pattern.Path != "" {
		testURL += pattern.Path
	}
	if pattern.Query != "" {
		if _, err := url.Parse(testURL); err == nil {
			testURL += "?" + pattern.Query
		}
	}

	// Create HTTP client with TLS skip verify (for self-signed certs)
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	// Parse the base URL to get host info
	parsedURL, err := url.Parse(req.TargetURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid target URL: %v", err), http.StatusBadRequest)
		return
	}

	// Create request with raw URL to prevent path normalization
	// This is important for path traversal tests
	httpReq, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create request: %v", err), http.StatusInternalServerError)
		return
	}

	// For path traversal attacks, we need to set the raw path to bypass Go's URL normalization
	if pattern.Path != "" {
		httpReq.URL.Opaque = "//" + parsedURL.Host + pattern.Path
		if pattern.Query != "" {
			httpReq.URL.RawQuery = pattern.Query
		}
	}

	// Add custom Host header if specified (for Docker internal testing)
	if req.HostHeader != "" {
		httpReq.Host = req.HostHeader
	}

	// Add custom header if specified
	if pattern.Header != "" && pattern.HeaderValue != "" {
		httpReq.Header.Set(pattern.Header, pattern.HeaderValue)
	}

	// Execute request and measure time
	start := time.Now()
	resp, err := client.Do(httpReq)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		// Request failed - could be blocked at network level
		result := WAFTestResult{
			AttackType:   req.AttackType,
			TestURL:      testURL,
			StatusCode:   0,
			Blocked:      true,
			ResponseTime: elapsed,
			Description:  pattern.Description + " (Connection failed - possibly blocked)",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}
	defer resp.Body.Close()

	// Read response body (limited)
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))

	// Determine if blocked
	blocked := resp.StatusCode == 403 || resp.StatusCode == 406 || resp.StatusCode == 429

	result := WAFTestResult{
		AttackType:   req.AttackType,
		TestURL:      testURL,
		StatusCode:   resp.StatusCode,
		Blocked:      blocked,
		ResponseTime: elapsed,
		Description:  pattern.Description,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *WAFTestHandler) ListPatterns(w http.ResponseWriter, r *http.Request) {
	type PatternInfo struct {
		ID          string `json:"id"`
		Category    string `json:"category"`
		Description string `json:"description"`
	}

	patterns := []PatternInfo{
		{ID: "sql_injection", Category: "SQL Injection", Description: "Classic OR-based bypass"},
		{ID: "sql_injection_union", Category: "SQL Injection", Description: "UNION-based attack"},
		{ID: "xss_script", Category: "XSS", Description: "Script tag injection"},
		{ID: "xss_event", Category: "XSS", Description: "Event handler injection"},
		{ID: "path_traversal", Category: "Path Traversal", Description: "Directory traversal attack"},
		{ID: "path_traversal_encoded", Category: "Path Traversal", Description: "URL encoded traversal"},
		{ID: "command_injection", Category: "Command Injection", Description: "Shell command injection"},
		{ID: "command_injection_pipe", Category: "Command Injection", Description: "Pipe command injection"},
		{ID: "scanner_sqlmap", Category: "Scanner Detection", Description: "SQLMap user agent"},
		{ID: "scanner_nikto", Category: "Scanner Detection", Description: "Nikto user agent"},
		{ID: "rce_php", Category: "RCE", Description: "PHP wrapper attack"},
		{ID: "protocol_attack", Category: "Protocol Attack", Description: "Host header injection"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(patterns)
}

func (h *WAFTestHandler) TestAll(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TargetURL  string `json:"target_url"`
		HostHeader string `json:"host_header,omitempty"` // Custom Host header for Docker internal testing
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TargetURL == "" {
		http.Error(w, "target_url is required", http.StatusBadRequest)
		return
	}

	// Parse the base URL to get host info
	parsedURL, err := url.Parse(req.TargetURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid target URL: %v", err), http.StatusBadRequest)
		return
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	var results []WAFTestResult

	for attackType, pattern := range attackPatterns {
		testURL := req.TargetURL
		if pattern.Path != "" {
			testURL += pattern.Path
		}
		if pattern.Query != "" {
			testURL += "?" + pattern.Query
		}

		httpReq, err := http.NewRequest("GET", testURL, nil)
		if err != nil {
			continue
		}

		// For path traversal attacks, we need to set the raw path to bypass Go's URL normalization
		if pattern.Path != "" {
			httpReq.URL.Opaque = "//" + parsedURL.Host + pattern.Path
			if pattern.Query != "" {
				httpReq.URL.RawQuery = pattern.Query
			}
		}

		// Add custom Host header if specified (for Docker internal testing)
		if req.HostHeader != "" {
			httpReq.Host = req.HostHeader
		}

		if pattern.Header != "" && pattern.HeaderValue != "" {
			httpReq.Header.Set(pattern.Header, pattern.HeaderValue)
		}

		start := time.Now()
		resp, err := client.Do(httpReq)
		elapsed := time.Since(start).Milliseconds()

		result := WAFTestResult{
			AttackType:   attackType,
			TestURL:      testURL,
			ResponseTime: elapsed,
			Description:  pattern.Description,
		}

		if err != nil {
			result.StatusCode = 0
			result.Blocked = true
		} else {
			io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
			resp.Body.Close()
			result.StatusCode = resp.StatusCode
			result.Blocked = resp.StatusCode == 403 || resp.StatusCode == 406 || resp.StatusCode == 429
		}

		results = append(results, result)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
