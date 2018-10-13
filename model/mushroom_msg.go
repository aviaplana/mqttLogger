package model

type MushroomMsg struct {
	Temperature    float32
	Humidity       float32
	IsLightOn      bool `json: light`
	IsHumidifierOn bool `json: humidifier`
	IsIntakeFanOn  bool `json: fan_intake`
	IsFlowFanOn    bool `json: fan_flow`
}

type MushroomMsgService interface {
	CreateMushroomMsg(u *MushroomMsg) error
	GetMushroomMsg(username string) (error, MushroomMsg)
}