package main

import (
	. "./mail"
	"encoding/json"
	"fmt"
	"github.com/aviaplana/mqttLogger/model"
	"github.com/aviaplana/mqttLogger/mongo"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

type Config struct {
	Email string
	MqttHost string
	MqttPort int
	MqttUser string
	MqttPassword string
	DbHost string
	DbPort int
	DbUser string
	DbPassword string
	DbName string
}

var mail Email
var mqttCli *client.Client

var mailSent = false

func main() {
	fmt.Println("Starting...")

	var config Config
	if err := getConfig(&config); err != nil {
		handleError(err)
	}

	mail = Email {ToEmail: config.Email}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill)

	db := connectDB(config.DbUser, config.DbPassword, config.DbHost, config.DbPort, config.DbName)
	defer db.Close()

	mqttCli = getMqttClient(config.MqttHost, config.MqttPort, config.MqttUser, config.MqttPassword)
	defer mqttCli.Terminate()

	// Subscribe to topics
	fridgeMsgService := mongo.NewFridgeMsgService(db.Copy(), config.DbName, "fridge")
	mushroomMsgService := mongo.NewMushroomMsgService(db.Copy(), config.DbName, "mushroom")
	if err := subscribeTopics(fridgeMsgService, mushroomMsgService); err != nil {
		handleError(err)
	}

	<-sigc

	// Disconnect the Network Connection.
	if err := mqttCli.Disconnect(); err != nil {
		handleError(err)
	}
}

func getConfig(configuration *Config) (error) {
	file, err := os.Open("config.json")

	if err != nil {  return err }

	decoder := json.NewDecoder(file)

	if err = decoder.Decode(&configuration); err != nil {
		return err
	}

	return nil
}

func subscribeTopics(fridgeService *mongo.FridgeMsgService, mushroomService *mongo.MushroomMsgService) (error) {
	return mqttCli.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			{
				TopicFilter: []byte("fridge/status"),
				QoS:         mqtt.QoS0,
				Handler: func(topicName, message []byte) {
					fmt.Println(string(topicName[:]), " -> ", string(message[:]))
					onReceiveFridgeMsg(message, fridgeService)
				},
			}, {
				TopicFilter: []byte("mushroom/status"),
				QoS:         mqtt.QoS0,
				Handler: func(topicName, message []byte) {
					fmt.Println(string(topicName[:]), " -> ", string(message[:]))
					onReceiveMushroomMsg(message, mushroomService)
				},
			},
		},
	})
}

func onReceiveMushroomMsg(message []byte, service *mongo.MushroomMsgService) {

	var mushroomMsg model.MushroomMsg
	json.Unmarshal(message, &mushroomMsg)

	if err := service.Create(&mushroomMsg); err != nil {
		handleError(err)
	}
}

func onReceiveFridgeMsg(message []byte, service *mongo.FridgeMsgService) {
	var fridgeRead model.FridgeMsg

	json.Unmarshal(message, &fridgeRead)
	if fridgeRead.Temperature == 0.0 {
		if !mailSent {
			sendMailMalfunction()
			mailSent = true
		}
	} else if mailSent {
		mailSent = false
		sendMailBackToWork()
	}

	if err := service.Create(&fridgeRead); err != nil {
		handleError(err)
	}
}

func sendMailBackToWork() {
	mail.Subject = "Fridge is working again"
	mail.Body = "Everything OK."
	mail.SendMail()
}

func sendMailMalfunction() {
	mail.Subject = "Fridge is malfunctioning"
	mail.Body = "Couldn't get sensor value."
	mail.SendMail()
}

func handleError(err error) {
	fmt.Println("Error", err)
	/*mail.Subject = "Fridge script is dead"
    mail.Body = "Cause: " + err.Error() 
    mail.SendMail()*/
	panic(err.Error())
}

func getMqttClient(host string, port int, user string, password string) (*client.Client) {
	mqttCli := client.New(&client.Options {
			ErrorHandler: func(err error) {
				fmt.Println(err)
			},
		})

	if err := connectMqtt(mqttCli, host, port, user, password); err != nil {
		handleError(err)
	} else {
		fmt.Println("Connected to the mqtt server.")
	}

	return mqttCli
}

func connectMqtt(mqttCli *client.Client, host string, port int, username string, password string) error {
	return mqttCli.Connect(&client.ConnectOptions {
		UserName:        []byte(username),
		Password:        []byte(password),
		Network:	"tcp",
		Address:	host + ":" + strconv.Itoa(port),
		ClientID:	[]byte("broker"),
	})
}

func connectDB(user string, password string, host string, port int, dbName string) (session *mongo.Session) {

	session, err := mongo.NewSession(user + ":" + password + "@" + host + ":" + strconv.Itoa(port))

	if err != nil {
		log.Fatalf("Unable to connect to mongo: %s", err)
	}

	return session
}
