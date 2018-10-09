package main

import (
	"fmt"
	"os"
	"os/signal"
	"encoding/json"
	"database/sql"
	. "./mail"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
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

type FridgeMsg struct {
	Temperature float32
	Humidity float32
	Pressure float32
	Compressor int
	Goal float32	`json: goal`
}

type MushroomMsg struct {
	Temperature float32
	Humidity float32
	isLightOn bool			`json: humidifier`
	isHumidifierOn bool		`json: humidifier`
	isIntakeFanOn bool		`json: fan_intake`
	isFanFlowOn bool		`json: fan_flow`
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

	if err := subscribeTopics(db); err != nil {
		handleError(err)
	}

	<-sigc

	// Disconnect the Network Connection.
	if err := mqttCli.Disconnect(); err != nil {
		handleError(err)
	}
}

func getConfig(configuration *Config) (error) {
	//filename is the path to the json config file
	file, err := os.Open("config.json")
	if err != nil {  return err }
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	if err != nil {  return err }
	return nil
}

func subscribeTopics(db *sql.DB) (error) {
	return mqttCli.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			{
				TopicFilter: []byte("mqttLogger/status"),
				QoS:         mqtt.QoS0,
				Handler: func(topicName, message []byte) {
					fmt.Println(string(topicName[:]), " -> ", string(message[:]))
					onReceiveFridgeMsg(message, db)
				},
			}, /*{
				TopicFilter: []byte("mushroom/status"),
				QoS:         mqtt.QoS0,
				Handler: func(topicName, message []byte) {
					onReceiveMushroomMsg(message, db)
				},
			},*/
		},
	})
}

func onReceiveMushroomMsg(message []byte, db *sql.DB) {
	mushroomStmt := getStmt(db, "INSERT INTO mqttLogger (temperature, humidity, pressure, compressor_on, goal_temperature) VALUES(?, ?, ?, ?, ? )")
	defer mushroomStmt.Close()

	var mushroomMsg MushroomMsg
	json.Unmarshal(message, &mushroomMsg)

	_, err := mushroomStmt.Exec(
		mushroomMsg.Temperature,
		mushroomMsg.Humidity,
		mushroomMsg.isLightOn,
		mushroomMsg.isHumidifierOn,
		mushroomMsg.isIntakeFanOn,
		mushroomMsg.isFanFlowOn)

	if err != nil {
		handleError(err)
	}
}

func onReceiveFridgeMsg(message []byte, db *sql.DB) {
	fridgeStmt := getStmt(db, "INSERT INTO mqttLogger (temperature, humidity, pressure, compressor_on, goal_temperature) VALUES(?, ?, ?, ?, ? )")
	defer fridgeStmt.Close()

	var fridgeRead FridgeMsg
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

	storeFridgeMsg(fridgeStmt, fridgeRead)
}

func storeFridgeMsg(stmtIns *sql.Stmt, fridgeRead FridgeMsg) {
	if fridgeRead.Temperature < 100 && fridgeRead.Humidity < 100 &&
		fridgeRead.Temperature > -100 && fridgeRead.Humidity > -100 {
		_, err := stmtIns.Exec(
			fridgeRead.Temperature,
			fridgeRead.Humidity,
			fridgeRead.Pressure,
			fridgeRead.Compressor,
			fridgeRead.Goal)
		if err != nil {
			handleError(err)
		}
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

func connectDB(user string, password string, host string, port int, dbName string) (*sql.DB) {
	db, err := sql.Open("mysql", user + ":" + password + "@tcp(" + host + ":" + strconv.Itoa(port) + ")/" + dbName )
	if err != nil {
		handleError(err)
	} else {
		fmt.Println("Connected to the database.")
	}

	return db
}

func getStmt(db *sql.DB, query string) (*sql.Stmt) {
	stmtIns, err := db.Prepare(query)

	if err != nil {
		handleError(err)
	}

	return stmtIns
}

