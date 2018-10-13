package mongo

import (
	"github.com/aviaplana/mqttLogger/model"
	"gopkg.in/mgo.v2/bson"
)

type MushroomMsgModel struct {
	Id          	bson.ObjectId `bson:"_id,omitempty"`
	Temperature 	float32
	Humidity 		float32
	IsLightOn 		bool
	IsHumidifierOn 	bool
	IsIntakeFanOn 	bool
	IsFlowFanOn 	bool
}

func newMushroomMsgModel(m *model.MushroomMsg) *MushroomMsgModel {
	return &MushroomMsgModel{
		Temperature: m.Temperature,
		Humidity: m.Humidity,
		IsLightOn: 		m.IsLightOn,
		IsHumidifierOn: m.IsHumidifierOn,
		IsIntakeFanOn: 	m.IsIntakeFanOn,
		IsFlowFanOn: 	m.IsFlowFanOn,
	}
}

func(m *MushroomMsgModel) toMushroomMsg() *model.MushroomMsg {
	return &model.MushroomMsg{
		Temperature:    m.Temperature,
		Humidity:       m.Humidity,
		IsLightOn:      m.IsLightOn,
		IsHumidifierOn: m.IsHumidifierOn,
		IsIntakeFanOn:  m.IsIntakeFanOn,
		IsFlowFanOn:    m.IsFlowFanOn,
	}
}