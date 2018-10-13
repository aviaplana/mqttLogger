package mongo

import (
	"github.com/aviaplana/mqttLogger/model"
	"gopkg.in/mgo.v2"
)

type MushroomMsgService struct {
	collection *mgo.Collection
}

func NewMushroomMsgService(session *Session, dbName string, collectionName string) *MushroomMsgService {
	collection := session.GetCollection(dbName, collectionName)
	//collection.EnsureIndex()
	return &MushroomMsgService{collection}
}

func(fms *MushroomMsgService) Create(m *model.MushroomMsg) error {
	mushroom := newMushroomMsgModel(m)
	return fms.collection.Insert(mushroom)
}

//func() GetByDate()