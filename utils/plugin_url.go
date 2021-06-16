package utils

import "fmt"

const domain = "http://sa.zhitingtech.com"

func DevicePluginURL(deviceID int, pluginName, model, name, token string) string {
	return fmt.Sprintf("%s/html/%s/html?device_id=%d&model=%s&name=%s&token=%s",
		domain, pluginName, deviceID, model, name, token)
}
