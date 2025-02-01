package useragent

import (
	"fmt"
	"net/http"

	"github.com/mileusna/useragent"
)

func GetDeviceName(r *http.Request) (deviceName string) {
	deviceNameMaxLen := 256
	ua := useragent.Parse(r.Header.Get("User-Agent"))
	if len(ua.Device) > 0 {
		deviceName = fmt.Sprintf("%v %v (%v)", ua.Name, ua.Version, ua.Device)
	} else {
		deviceName = fmt.Sprintf("%v %v", ua.Name, ua.Version)
	}

	if len(deviceName) > deviceNameMaxLen {
		deviceName = deviceName[:deviceNameMaxLen]
	}

	return
}

func GetDeviceType(r *http.Request) (t string) {
	ua := useragent.Parse(r.Header.Get("User-Agent"))

	switch {
	case ua.Mobile:
		t = "Mobile"
	case ua.Tablet:
		t = "Tablet"
	case ua.Desktop:
		t = "Desktop"
	case ua.Bot:
		t = "Bot"
	default:
		t = "unknown"
	}

	return
}

func GetDeviceOS(r *http.Request) string {
	deviceOSMaxLen := 64
	ua := useragent.Parse(r.Header.Get("User-Agent"))
	deviceOS := ua.OS + " " + ua.OSVersion

	if len(deviceOS) > deviceOSMaxLen {
		deviceOS = deviceOS[:deviceOSMaxLen]
	}

	return deviceOS
}
