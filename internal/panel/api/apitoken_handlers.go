package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/panel/service"
)

// APITokenHandlers manages API token CRUD endpoints.
type APITokenHandlers struct {
	Svc *service.APITokenService
}

// ListAPITokens returns all tokens (secrets excluded).
func (h *APITokenHandlers) ListAPITokens(c echo.Context) error {
	tokens, err := h.Svc.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"tokens": tokens})
}

type createTokenRequest struct {
	Name string `json:"name"`
}

// CreateAPIToken generates a new token. The raw secret is returned once.
func (h *APITokenHandlers) CreateAPIToken(c echo.Context) error {
	var req createTokenRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}
	result, err := h.Svc.Create(c.Request().Context(), claims.AdminID, req.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, result)
}

// DeleteAPIToken revokes a token.
func (h *APITokenHandlers) DeleteAPIToken(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Svc.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "delete failed")
	}
	return c.NoContent(http.StatusNoContent)
}
