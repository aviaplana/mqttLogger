package model

type FridgeMsg struct {
	Temperature float32	`json: temperature`
	Humidity float32	`json: humidity`
	Pressure float32	`json: pressure`
	Compressor bool		`json: compressor`
	Goal float32		`json: goal`
}

type FridgeMsgService interface {
	CreateFridgeMsg(u *FridgeMsg) error
	GetFridgeMsg(username string) (error, FridgeMsg)
}