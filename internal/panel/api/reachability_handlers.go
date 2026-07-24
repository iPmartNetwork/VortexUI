package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// ReachabilityHandler provides Iran TCP reachability checks via check-host.net.
type ReachabilityHandler struct {
	client *http.Client
}

// NewReachabilityHandler creates a ReachabilityHandler with a default HTTP client.
func NewReachabilityHandler() *ReachabilityHandler {
	return &ReachabilityHandler{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Register mounts reachability routes on the given Echo group.
func (h *ReachabilityHandler) Register(g *echo.Group) {
	r := g.Group("/reachability")
	r.POST("/check", h.Check)
}

// reachabilityRequest is the request body for POST /api/v2/reachability/check.
type reachabilityRequest struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// reachabilityNode represents a single check-host.net node result.
type reachabilityNode struct {
	Node   string `json:"node"`
	Up     bool   `json:"up"`
	Detail string `json:"detail"`
}

// reachabilityResponse is the final response for the reachability check.
type reachabilityResponse struct {
	Verdict string             `json:"verdict"` // "reachable", "unreachable", "partial"
	OK      int                `json:"ok"`
	Total   int                `json:"total"`
	Nodes   []reachabilityNode `json:"nodes"`
}

// iranNodes are the check-host.net Iran node identifiers.
var iranNodes = []string{
	"ir1.node.check-host.net",
	"ir2.node.check-host.net",
}

// Check handles POST /api/v2/reachability/check.
// Takes {host, port}, calls check-host.net HTTP API to test TCP reachability
// from Iran nodes. Returns verdict with per-node results.
func (h *ReachabilityHandler) Check(c echo.Context) error {
	var req reachabilityRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Host == "" || req.Port <= 0 || req.Port > 65535 {
		return echo.NewHTTPError(http.StatusBadRequest, "host and valid port (1-65535) are required")
	}

	ctx := c.Request().Context()

	// Step 1: Initiate TCP check on check-host.net.
	target := fmt.Sprintf("%s:%d", req.Host, req.Port)
	checkURL := fmt.Sprintf("https://check-host.net/check-tcp?host=%s", target)
	for i, node := range iranNodes {
		sep := "&"
		if i == 0 {
			sep = "&"
		}
		checkURL += fmt.Sprintf("%snode=%s", sep, node)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, checkURL, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create request")
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := h.client.Do(httpReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "check-host.net unreachable: "+err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("check-host.net returned %d: %s", resp.StatusCode, string(body)))
	}

	// Parse the initial response to get request_id.
	var initResp struct {
		RequestID string `json:"request_id"`
		OK        int    `json:"ok"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&initResp); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to parse check-host response")
	}
	if initResp.RequestID == "" {
		return echo.NewHTTPError(http.StatusBadGateway, "check-host.net did not return request_id")
	}

	// Step 2: Wait 5 seconds for results to be available.
	select {
	case <-time.After(5 * time.Second):
	case <-ctx.Done():
		return echo.NewHTTPError(http.StatusRequestTimeout, "request cancelled")
	}

	// Step 3: Fetch results.
	resultURL := fmt.Sprintf("https://check-host.net/check-result/%s", initResp.RequestID)
	resultReq, err := http.NewRequestWithContext(ctx, http.MethodGet, resultURL, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create result request")
	}
	resultReq.Header.Set("Accept", "application/json")

	resultResp, err := h.client.Do(resultReq)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "failed to fetch results: "+err.Error())
	}
	defer resultResp.Body.Close()

	// Parse results: map of node -> [[status, latency_or_error]]
	var results map[string]json.RawMessage
	if err := json.NewDecoder(resultResp.Body).Decode(&results); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to parse results")
	}

	// Step 4: Interpret results and build response.
	var nodes []reachabilityNode
	okCount := 0
	total := 0

	for _, nodeName := range iranNodes {
		total++
		raw, exists := results[nodeName]
		if !exists || raw == nil || string(raw) == "null" {
			nodes = append(nodes, reachabilityNode{Node: nodeName, Up: false, Detail: "no result"})
			continue
		}

		// check-host returns [[{"address":"...","time":0.123}]] for success
		// or [[{"error":"..."}]] for failure, wrapped as arrays.
		var entries [][]interface{}
		if err := json.Unmarshal(raw, &entries); err != nil {
			nodes = append(nodes, reachabilityNode{Node: nodeName, Up: false, Detail: "parse error"})
			continue
		}

		if len(entries) > 0 && len(entries[0]) > 0 {
			// Try to determine if the connection succeeded.
			entry := entries[0][0]
			switch v := entry.(type) {
			case map[string]interface{}:
				if errMsg, hasErr := v["error"]; hasErr {
					nodes = append(nodes, reachabilityNode{Node: nodeName, Up: false, Detail: fmt.Sprintf("%v", errMsg)})
				} else {
					okCount++
					latency := ""
					if t, ok := v["time"]; ok {
						latency = fmt.Sprintf("%.3fms", t)
					}
					nodes = append(nodes, reachabilityNode{Node: nodeName, Up: true, Detail: latency})
				}
			default:
				// Numeric result often means success (latency in seconds).
				okCount++
				nodes = append(nodes, reachabilityNode{Node: nodeName, Up: true, Detail: fmt.Sprintf("%v", v)})
			}
		} else {
			nodes = append(nodes, reachabilityNode{Node: nodeName, Up: false, Detail: "empty result"})
		}
	}

	verdict := "unreachable"
	if okCount == total {
		verdict = "reachable"
	} else if okCount > 0 {
		verdict = "partial"
	}

	return c.JSON(http.StatusOK, reachabilityResponse{
		Verdict: verdict,
		OK:      okCount,
		Total:   total,
		Nodes:   nodes,
	})
}
