package assets

import "embed"

//go:embed icons/*
var icons embed.FS

type Icon int

const (
	ShieldOff Icon = iota
	ShieldOn
	Umbrella
)

func (i Icon) fileName() string {
	switch i {
	case ShieldOn:
		return "icons/shield.png"
	case Umbrella:
		return "icons/umbrella.png"
	default:
		return "icons/shield_off.png"
	}
}

func GetIcon(file Icon) []byte {
	data, err := icons.ReadFile(file.fileName())
	if err != nil {
		panic(err)
	}
	return data
}
