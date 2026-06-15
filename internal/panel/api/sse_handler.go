package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/events"
)

// EventStream is the subset of the event bus the SSE endpoint needs.
type EventStream interface {
	Subscribe(buffer int) <-chan events.Event
	Unsubscribe(ch <-chan events.Event)
}

// StreamEvents pushes domain events to the browser over Server-Sent Events. The
// frontend uses it to refresh views the instant something changes (a node goes
// down, a user is limited, sharing is detected) instead of polling. Auth is the
// standard bearer flow, with the token allowed as ?access_token= since the
// EventSource API cannot set headers.
func (h *Handlers) StreamEvents(c echo.Context) error {
	if h.Events == nil {
		return echo.NewHTTPError(http.StatusNotFound, "live events not enabled")
	}
	r := c.Response()
	r.Header().Set(echo.HeaderContentType, "text/event-stream")
	r.Header().Set(echo.HeaderCacheControl, "no-cache")
	r.Header().Set(echo.HeaderConnection, "keep-alive")
	r.Header().Set("X-Accel-Buffering", "no") // disable proxy buffering
	r.WriteHeader(http.StatusOK)
	r.Flush()

	ch := h.Events.Subscribe(64)
	defer h.Events.Unsubscribe(ch)

	ctx := c.Request().Context()
	// Greeting so the client knows the stream is live.
	fmt.Fprint(r, "event: ready\ndata: {}\n\n")
	r.Flush()

	for {
		select {
		case <-ctx.Done():
			return nil
		case ev, ok := <-ch:
			if !ok {
				return nil
			}
			payload, err := json.Marshal(ev)
			if err != nil {
				continue
			}
			if _, err := fmt.Fprintf(r, "event: %s\ndata: %s\n\n", ev.Type, payload); err != nil {
				return nil
			}
			r.Flush()
		}
	}
}
