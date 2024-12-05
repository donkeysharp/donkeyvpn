package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/donkeysharp/donkeyvpn/internal/aws"
	"github.com/donkeysharp/donkeyvpn/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type VPNHandler struct {
	WebhookSecret  string
	InstancesTable *aws.DynamoDB
}

func (h *VPNHandler) Handle(c echo.Context) error {
	contentRaw, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Error("could not load body reader")
		return c.String(http.StatusAccepted, GenericErrorMessage)
	}

	if len(contentRaw) == 0 {
		contentRaw = []byte("Empty")
	}
	c.Response().Header().Add("Content-type", "application/json")
	content := string(contentRaw)
	requestToken := c.Request().Header.Get("x-api-key")
	if strings.Compare(requestToken, h.WebhookSecret) != 0 {
		log.Error("received a missing or invalid webhook secret")
		c.Response().Header().Add("content-type", "text/plain")
		return c.String(http.StatusAccepted, GenericErrorMessage)
	}

	log.Infof("Body content: %s", content)

	var tmp interface{}
	err = json.Unmarshal(contentRaw, &tmp)
	if err != nil {
		log.Error("error unmarshalling request body")
		return c.String(http.StatusAccepted, GenericErrorMessage)
	}
	vpnInfo := tmp.(map[string]interface{})
	hostname, ok := vpnInfo["hostname"].(string)
	if !ok {
		return c.String(http.StatusBadRequest, "hostname and publicIP fields are expected")
	}
	publicIP, ok := vpnInfo["publicIP"].(string)
	if !ok {
		return c.String(http.StatusBadRequest, "hostname and publicIP fields are expected")
	}
	instance := models.NewVPNInstance(hostname, publicIP)
	created, err := h.InstancesTable.CreateRecord(instance)
	if err != nil {
		log.Errorf("error registering instance %v", err)
		return c.String(http.StatusInternalServerError, "Error registering the instance")
	}

	if !created {
		log.Error("The VPN Instance was not created")
		return c.String(http.StatusInternalServerError, "The VPN instance was not created")
	}

	return c.String(http.StatusAccepted, "OK")
}
