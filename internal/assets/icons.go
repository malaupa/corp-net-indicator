package assets

import "embed"

//go:embed icons/*
var icons embed.FS

type Icon string

const (
	ShieldOff  Icon = "icons/shield_off.png"
	ShieldOn   Icon = "icons/shield.png"
	Umbrella   Icon = "icons/umbrella.png"
	Connect    Icon = "icons/connect.png"
	Disconnect Icon = "icons/disconnect.png"
	Status     Icon = "icons/activity.png"
)

func GetIcon(file Icon) []byte {
	data, err := icons.ReadFile(string(file))
	if err != nil {
		panic(err)
	}
	return data
}
