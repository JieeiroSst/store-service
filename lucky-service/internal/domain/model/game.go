package model

import "strings"

type GameType string

const (
	GameWheel       GameType = "wheel"
	GameScratchCard GameType = "scratch_card"
	GameGiftBox     GameType = "gift_box"
	GameShake       GameType = "shake"
	GameFlipCard    GameType = "flip_card"
	GameSlotMachine GameType = "slot_machine"
	GameCapsule     GameType = "capsule"
	GameCheckin     GameType = "checkin"
)

var Games = []GameType{GameWheel, GameScratchCard, GameGiftBox, GameShake, GameFlipCard, GameSlotMachine, GameCapsule, GameCheckin}

func ParseGame(s string) (GameType, error) {
	g := GameType(strings.ToLower(strings.TrimSpace(s)))
	if !g.Valid() {
		return "", invalid("game_type", "must be one of wheel, scratch_card, gift_box, shake, flip_card, slot_machine, capsule, checkin")
	}
	return g, nil
}

func (g GameType) Valid() bool {
	for _, v := range Games {
		if g == v {
			return true
		}
	}
	return false
}

func (g GameType) Instant() bool { return g.Valid() && g != GameCheckin }

func (g GameType) NeedsChoice() bool { return g == GameGiftBox || g == GameFlipCard }
