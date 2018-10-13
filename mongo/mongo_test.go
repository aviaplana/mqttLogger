package mongo_test

import (
	"github.com/aviaplana/mqttLogger/model"
	"github.com/aviaplana/mqttLogger/mongo"
	"log"
	"testing"
)

const (
	mongoUrl = "10.0.0.6:27017"
	dbName = "test_db"
	fridgeCollectionName = "fridge"
)

func Test_FridgeMsgService(t *testing.T) {
	t.Run("CreateFridgeMsg", createFridgeMsg_should_insert_msg_into_mongo)
}

func createFridgeMsg_should_insert_msg_into_mongo(t *testing.T) {
	session, err := mongo.NewSession(mongoUrl)

	if err != nil {
		log.Fatalf("Unable to connect to mongo: %s", err)
	}
	defer func() {
		session.DropDatabase(dbName)
		session.Close()
	}()
	fridgeMsgService := mongo.NewFridgeMsgService(session.Copy(), dbName, fridgeCollectionName)

	fridgeMsg := model.FridgeMsg{
		Temperature: 10.0,
		Humidity:	11.0,
		Pressure:	12.0,
		Compressor:	true,
		Goal:		13.3,
	}

	err = fridgeMsgService.Create(&fridgeMsg)

	if err != nil {
		t.Error("Unable to create fridge msg: %s", err)
	}

	var results []model.FridgeMsg
	session.GetCollection(dbName, fridgeCollectionName).Find(nil).All(&results)


	if count := len(results); count != 1 {
		t.Error("Incorrect number of results. Expected `1` got: %i", count)
	}

	if results[0].Temperature != fridgeMsg.Temperature {
		t.Error("Incorrect temperature. Expected %f, got %f", fridgeMsg.Temperature, results[0].Temperature)
	}



}