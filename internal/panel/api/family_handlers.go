package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/panel/service"
)

// FamilyHandlers serves family/group subscription endpoints.
type FamilyHandlers struct {
	Family *service.FamilyService
}

type createGroupRequest struct {
	Name        string `json:"name"`
	OwnerID     string `json:"owner_id"`
	DataLimit   int64  `json:"data_limit"`
	MaxMembers  int    `json:"max_members"`
	MemberQuota int64  `json:"member_quota"`
}

// CreateGroup creates a new family group.
func (h *FamilyHandlers) CreateGroup(c echo.Context) error {
	var req createGroupRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid owner_id")
	}
	g, err := h.Family.CreateGroup(c.Request().Context(), service.CreateGroupInput{
		Name:        req.Name,
		OwnerID:     ownerID,
		DataLimit:   req.DataLimit,
		MaxMembers:  req.MaxMembers,
		MemberQuota: req.MemberQuota,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"group": g})
}

// ListGroups lists all family groups.
func (h *FamilyHandlers) ListGroups(c echo.Context) error {
	groups, total, err := h.Family.ListGroups(c.Request().Context(), 50, 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"groups": groups, "total": total})
}

// GetGroup returns a single group with members.
func (h *FamilyHandlers) GetGroup(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	g, err := h.Family.GetGroup(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"group": g})
}

type addMemberRequest struct {
	UserID string `json:"user_id"`
	Label  string `json:"label"`
}

// AddMember adds a user to a group.
func (h *FamilyHandlers) AddMember(c echo.Context) error {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req addMemberRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id")
	}
	m, err := h.Family.AddMember(c.Request().Context(), groupID, userID, req.Label)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"member": m})
}

// RemoveMember removes a user from a group.
func (h *FamilyHandlers) RemoveMember(c echo.Context) error {
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	userID, err := uuid.Parse(c.Param("uid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user_id")
	}
	if err := h.Family.RemoveMember(c.Request().Context(), groupID, userID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// DeleteGroup removes a family group.
func (h *FamilyHandlers) DeleteGroup(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Family.DeleteGroup(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "delete failed")
	}
	return c.NoContent(http.StatusNoContent)
}
