package handler

import (
	"fmt"
	"net/http"

	"github.com/donkeysharp/donkeyvpn/internal/models"
	"github.com/donkeysharp/donkeyvpn/internal/service"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type JSONObj map[string]interface{}

type VPNHandler struct {
	WebhookSecret  string
	VPNSvc         *service.VPNService
	TelegramClient *telegram.Client
}

type NotificationRequest struct {
	Id         string
	Hostname   string `json:"hostname"`
	Port       string `json:"port"`
	Status     string `json:"status"`
	InstanceId string `json:"instanceId"`
}

func (r *NotificationRequest) ToModel() interface{} {
	return models.VPNInstance{
		Id:         r.Id,
		Hostname:   r.Hostname,
		Port:       r.Port,
		Status:     r.Status,
		InstanceId: r.InstanceId,
	}
}

func (h *VPNHandler) GetPendingId(c echo.Context) error {
	instances, err := h.VPNSvc.ListPending()
	if err != nil {
		log.Errorf("Error retrieving pending instances: %v", err.Error())
		return c.String(http.StatusInternalServerError, GenericErrorMessage)
	}
	return c.JSON(http.StatusAccepted, instances)
}

func (h *VPNHandler) Notify(c echo.Context) error {
	vpnId := c.Param("vpnId")
	var req *NotificationRequest = new(NotificationRequest)
	if err := c.Bind(&req); err != nil {
		log.Errorf("Could not process %v", err.Error())
		return c.JSON(http.StatusBadRequest, JSONObj{"error": GenericErrorMessage})
	}
	req.Id = vpnId
	log.Infof("	 vpnId: %s hostname: %s status: %s", req.Id, req.Hostname, req.Status)
	instance, err := h.VPNSvc.Update(req.ToModel().(models.VPNInstance))
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

	if instance.ChatId != "" {
		chat := &telegram.Chat{
			ChatId: instance.ChatIdValue(),
		}
		message := fmt.Sprintf("VPN Instance with id %v provisioned successfully.", instance.Id)
		h.TelegramClient.SendMessage(message, chat)
	}

	log.Infof("NotificationRequest: %v ", req)
	return c.JSON(http.StatusAccepted, JSONObj{
		"message": "VPN instance registered successfully",
	})
}

func (h *VPNHandler) Ping(c echo.Context) error {
	log.Info("Ping event")
	return c.JSON(http.StatusAccepted, JSONObj{
		"message": "Pong from donkeyvpn",
	})
}
