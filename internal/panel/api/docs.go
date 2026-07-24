package api

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed swagger-ui
var swaggerUI embed.FS

// DocsHandler serves embedded Swagger UI and OpenAPI spec.
type DocsHandler struct {
	specPath string // path to openapi.yaml relative to working directory
}

// NewDocsHandler creates the docs handler.
func NewDocsHandler(specPath string) *DocsHandler {
	if specPath == "" {
		specPath = "docs/openapi.yaml"
	}
	return &DocsHandler{specPath: specPath}
}

// Register mounts the API documentation routes.
func (h *DocsHandler) Register(g *echo.Group) {
	// Serve OpenAPI spec
	g.GET("/docs/openapi.yaml", h.ServeSpec)
	g.GET("/docs/openapi.json", h.ServeSpec)

	// Serve Swagger UI (static files)
	subFS, err := fs.Sub(swaggerUI, "swagger-ui")
	if err == nil {
		g.GET("/docs/*", echo.WrapHandler(http.StripPrefix("/api/v2/docs/", http.FileServer(http.FS(subFS)))))
	}

	// Redirect /docs to /docs/index.html
	g.GET("/docs", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/api/v2/docs/index.html")
	})
}

// ServeSpec serves the OpenAPI specification file.
func (h *DocsHandler) ServeSpec(c echo.Context) error {
	return c.File(h.specPath)
}
