package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"nginx-proxy-guard/internal/model"
	"nginx-proxy-guard/internal/service"
	"nginx-proxy-guard/internal/util"
)

type CertificateHandler struct {
	service *service.CertificateService
	audit   *service.AuditService
}

func NewCertificateHandler(svc *service.CertificateService, audit *service.AuditService) *CertificateHandler {
	return &CertificateHandler{service: svc, audit: audit}
}

// List handles GET /api/v1/certificates
func (h *CertificateHandler) List(c echo.Context) error {
	page, perPage := ParsePaginationParams(c)

	response, err := h.service.List(c.Request().Context(), page, perPage)
	if err != nil {
		return databaseError(c, "list certificates", err)
	}

	return c.JSON(http.StatusOK, response)
}

// Get handles GET /api/v1/certificates/:id
func (h *CertificateHandler) Get(c echo.Context) error {
	id := c.Param("id")

	cert, err := h.service.GetByID(c.Request().Context(), id)
	if err != nil {
		return databaseError(c, "get certificate", err)
	}

	if cert == nil {
		return notFoundError(c, "Certificate")
	}

	return c.JSON(http.StatusOK, cert.ToWithDetails())
}

// Create handles POST /api/v1/certificates
func (h *CertificateHandler) Create(c echo.Context) error {
	var req model.CreateCertificateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	if len(req.DomainNames) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "domain_names is required",
		})
	}

	cert, err := h.service.Create(c.Request().Context(), &req)
	if err != nil {
		return internalError(c, "create certificate", err)
	}

	// Audit log
	auditCtx := service.ContextWithAudit(c.Request().Context(), c)
	h.audit.LogCertificateCreate(auditCtx, req.DomainNames, "letsencrypt")

	return c.JSON(http.StatusCreated, cert)
}

// Upload handles POST /api/v1/certificates/upload
func (h *CertificateHandler) Upload(c echo.Context) error {
	var req model.UploadCertificateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	if req.CertificatePEM == "" || req.PrivateKeyPEM == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "certificate_pem and private_key_pem are required",
		})
	}

	cert, err := h.service.UploadCustom(c.Request().Context(), &req)
	if err != nil {
		return badRequestError(c, "Invalid certificate or key format")
	}

	// Audit log
	auditCtx := service.ContextWithAudit(c.Request().Context(), c)
	h.audit.LogCertificateCreate(auditCtx, cert.DomainNames, "custom")

	return c.JSON(http.StatusCreated, cert)
}

// Delete handles DELETE /api/v1/certificates/:id
func (h *CertificateHandler) Delete(c echo.Context) error {
	id := c.Param("id")

	// Get certificate info before deletion for audit
	cert, _ := h.service.GetByID(c.Request().Context(), id)

	err := h.service.Delete(c.Request().Context(), id)
	if err != nil {
		if err == model.ErrNotFound {
			return notFoundError(c, "Certificate")
		}
		return internalError(c, "delete certificate", err)
	}

	// Audit log
	if cert != nil {
		auditCtx := service.ContextWithAudit(c.Request().Context(), c)
		h.audit.LogCertificateDelete(auditCtx, cert.DomainNames)
	}

	return c.NoContent(http.StatusNoContent)
}

// Renew handles POST /api/v1/certificates/:id/renew
func (h *CertificateHandler) Renew(c echo.Context) error {
	id := c.Param("id")

	// Get certificate info for audit
	cert, _ := h.service.GetByID(c.Request().Context(), id)

	err := h.service.Renew(c.Request().Context(), id)
	if err != nil {
		if err == model.ErrNotFound {
			return notFoundError(c, "Certificate")
		}
		return internalError(c, "renew certificate", err)
	}

	// Audit log
	if cert != nil {
		auditCtx := service.ContextWithAudit(c.Request().Context(), c)
		h.audit.LogCertificateRenewed(auditCtx, cert.DomainNames)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "certificate renewal initiated",
	})
}

// GetExpiring handles GET /api/v1/certificates/expiring
func (h *CertificateHandler) GetExpiring(c echo.Context) error {
	days := util.ParseIntParam(c, "days", 30, 1, 365)

	certs, err := h.service.GetExpiringSoon(c.Request().Context(), days)
	if err != nil {
		return databaseError(c, "get expiring certificates", err)
	}

	// Convert to details
	result := make([]model.CertificateWithDetails, len(certs))
	for i, cert := range certs {
		result[i] = cert.ToWithDetails()
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  result,
		"total": len(result),
	})
}

// GetLogs handles GET /api/v1/certificates/:id/logs
// Returns the real-time logs for a certificate issuance process
func (h *CertificateHandler) GetLogs(c echo.Context) error {
	id := c.Param("id")

	logs, err := h.service.GetCertLogs(c.Request().Context(), id)
	if err != nil {
		if err == model.ErrNotFound {
			return notFoundError(c, "Certificate")
		}
		return internalError(c, "get certificate logs", err)
	}

	return c.JSON(http.StatusOK, logs)
}

// ListHistory handles GET /api/v1/certificates/history
// Returns paginated certificate history (issuance, renewal, errors)
func (h *CertificateHandler) ListHistory(c echo.Context) error {
	page, perPage := ParsePaginationParams(c)
	certificateID := c.QueryParam("certificate_id")

	response, err := h.service.ListHistory(c.Request().Context(), page, perPage, certificateID)
	if err != nil {
		return databaseError(c, "list certificate history", err)
	}

	return c.JSON(http.StatusOK, response)
}
