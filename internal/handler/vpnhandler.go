package handler

import (
	"net/http"

	"github.com/donkeysharp/donkeyvpn/internal/models"
	"github.com/donkeysharp/donkeyvpn/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type JSONObj map[string]interface{}

type VPNHandler struct {
	WebhookSecret string
	VPNSvc        *service.VPNService
}

type NotificationRequest struct {
	Id       string
	Hostname string `json:"hostname"`
	Port     string `json:"port"`
	Status   string `json:"status"`
}

func (r *NotificationRequest) ToModel() interface{} {
	return models.VPNInstance{
		Id:       r.Id,
		Hostname: r.Hostname,
		Port:     r.Port,
		Status:   r.Status,
	}
}

func (h *VPNHandler) NextId(c echo.Context) error {
	nextId, err := h.VPNSvc.NextId()
	if err != nil {
		log.Errorf("Could not get next id: %v", err.Error())
		return c.String(http.StatusInternalServerError, GenericErrorMessage)
	}
	response := map[string]string{
		"nextId": nextId,
	}
	return c.JSON(http.StatusAccepted, response)
}

func (h *VPNHandler) Get(c echo.Context) error {
	vpnId := c.Param("vpnId")
	instance, err := h.VPNSvc.Get(vpnId)
	if err != nil {
		if err == service.ErrVPNInstanceNotFound {
			return c.JSON(http.StatusNotFound, JSONObj{
				"message": "VPN instance not found",
			})
		}
		log.Errorf("Failed to retrieve VPN instance: %v", err.Error())
		return c.JSON(http.StatusInternalServerError, JSONObj{
			"message": GenericErrorMessage,
		})
	}
	return c.JSON(http.StatusOK, instance)
}

func (h *VPNHandler) Notify(c echo.Context) error {
	vpnId := c.Param("vpnId")

	var req *NotificationRequest = new(NotificationRequest)
	if err := c.Bind(&req); err != nil {
		log.Errorf("Could not process %v", err.Error())
		return c.JSON(http.StatusBadRequest, JSONObj{"error": GenericErrorMessage})
	}
	req.Id = vpnId
	log.Infof("Registering vpnId: %s hostname: %s status: %s", req.Id, req.Hostname, req.Status)
	err := h.VPNSvc.Update(req.ToModel().(models.VPNInstance))
	if err != nil {
		if err == service.ErrVPNInstanceNotFound {
			return c.JSON(http.StatusNotFound, JSONObj{
				"message": "VPN instance not found",
			})
		}
		log.Errorf("Failed processing request: %v", err.Error())
		return c.JSON(http.StatusInternalServerError, JSONObj{
			"message": GenericErrorMessage,
		})
	}

	log.Infof("NotificationRequest: %v ", req)
	return c.JSON(http.StatusAccepted, JSONObj{
		"message": "VPN instance registered",
	})
}
