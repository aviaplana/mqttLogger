package mongo

import (
	"github.com/aviaplana/mqttLogger/model"
	"gopkg.in/mgo.v2"
)

type FridgeMsgService struct {
	collection *mgo.Collection
}

func NewFridgeMsgService(session *Session, dbName string, collectionName string) *FridgeMsgService {
	collection := session.GetCollection(dbName, collectionName)
	//collection.EnsureIndex()
	return &FridgeMsgService{collection}
}

func(fms *FridgeMsgService) Create(f *model.FridgeMsg) error {
	fridge := newFridgeMsgModel(f)
	return fms.collection.Insert(fridge)
}

//func() GetByDate()