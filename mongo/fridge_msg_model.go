package mongo

import (
	"github.com/aviaplana/mqttLogger/model"
	"gopkg.in/mgo.v2/bson"
)

type FridgeMsgModel struct {
	Id          bson.ObjectId `bson:"_id,omitempty"`
	Temperature float32
	Humidity    float32
	Pressure    float32
	Compressor  bool
	Goal        float32
}

func newFridgeMsgModel(f *model.FridgeMsg) *FridgeMsgModel {
	return &FridgeMsgModel{
		Temperature: f.Temperature,
		Humidity: f.Humidity,
		Pressure: f.Pressure,
		Compressor: f.Compressor,
		Goal: f.Goal }
}

func(f *FridgeMsgModel) toFridgeMsg() *model.FridgeMsg {
	return &model.FridgeMsg{
		Temperature: f.Temperature,
		Humidity: f.Humidity,
		Pressure: f.Pressure,
		Compressor: f.Compressor,
		Goal: f.Goal }
}