package mahjong

import (
	"encoding/json"
	"fmt"
)

var Condition = map[string]int{
	"Eat":    0,
	"Pon":    1,
	"Gon":    2,
	"OnGon":  3,
	"PonGon": 4,
	"Hu":     5,
	"Zimo":   6,
	"Tin":    7,
}

type GossipLog struct {
	Name     string       `bson:"Name,omitempty"`
	Actions  []ActionInfo `bson:"Actions,omitempty"`
	Sentence string       `bson:"Sentence,omitempty"`
}

type Sentence struct {
	Sentence string `bson:"Sentence,omitempty"`
}
type GossipInfo struct {
	CurAction   string
	Id          int
	Hand        SuitSet
	ThrowTimes  int
	StepsToHu   int
	IsTing      bool
	OtherTing   bool
	SafeBefore  []string
	Actions     []ActionInfo
	DeadCards   SuitSet
	PlayerNames []string
}
type ThrowCardInfo struct {
	Id          int
	Card        string
	PlayerNames []string
}

func (room Room) Record(action string, card Tile, currentIdx int, actionIdx int) {

	if action == "Eat" || action == "Pon" || action == "Gon" || action == "Hu" {
		from := (actionIdx - currentIdx + 4) % 4
		room.Players[currentIdx].Actions = append(room.Players[currentIdx].Actions, ActionInfo{room.Turn, "Ba" + action, from, card.ToString()})
		from = (currentIdx - actionIdx + 4) % 4
		room.Players[actionIdx].Actions = append(room.Players[actionIdx].Actions, ActionInfo{room.Turn, action, from, card.ToString()})

		if sub := (actionIdx - currentIdx + 4) % 4; sub > 1 { //打牌的下家被跳過
			for i := 1; i < sub; i++ {
				room.Players[(currentIdx+i)%4].Actions = append(room.Players[(currentIdx+i)%4].Actions, ActionInfo{room.Turn, "PassDraw", -1, ""})
			}
		}
	} else if action == "Throw" || action == "Ting" || action == "OnGon" || action == "PonGon" || action == "Hu" || action == "CantEat" {
		room.Players[actionIdx].Actions = append(room.Players[actionIdx].Actions, ActionInfo{room.Turn, action, -1, card.ToString()})
	} else if action == "Draw" {
		room.Players[actionIdx].Actions = append(room.Players[actionIdx].Actions, ActionInfo{room.Turn, action, -1, card.ToString()})
		if room.Players[actionIdx].StepsToHu > room.Players[actionIdx].Hand.CountStepsToHu() {
			room.Players[actionIdx].Actions = append(room.Players[actionIdx].Actions, ActionInfo{room.Turn, "SubTingNum", -1, card.ToString()})
		}
	}

	room.GossipTalk(action, card.ToString(), actionIdx)
}

func (room Room) CountDeadCard() SuitSet {
	amount := SuitSet{}

	for i := 0; i < 5; i++ {
		for j := 0; j < 4; j++ {
			amount[i] += room.Players[j].Flowers[i]
			amount[i] += room.Players[j].DiscardTiles[i]
			amount[i] += room.Players[j].PonTiles[i] * 3
			amount[i] += room.Players[j].GonTiles[i] * 4
		}
	}
	for i := 0; i < 4; i++ {
		for _, eat := range room.Players[i].EatTiles {
			suit := eat.First.Suit
			value := eat.First.Value
			amount[suit].Add(value)
			value += 1
			amount[suit].Add(value)
			value += 1
			amount[suit].Add(value)
		}
	}

	return amount
}

func (room Room) GossipTalk(action string, card string, id int) {
	player := room.Players[id]
	var name []string
	for _, player := range room.Players {
		name = append(name, PlayerList[FindPlayerByUUID(player.UUID)].Name)
	}
	if action == "Throw" {
		throwCardInfo := &ThrowCardInfo{id, card, name}
		info, err := json.Marshal(throwCardInfo)
		if err != nil {
			fmt.Println(err)
		}
		game.GossipDealer.Emit("throwCardInfo", string(info), func(pid int, sentence string) {
			fmt.Println(pid, sentence)
			room.BroadcastCoversation(pid, sentence)
		})
	}

	gossipInfo := &GossipInfo{action, id, player.Hand, player.ThrowTimes, player.StepsToHu,
		player.IsTing, player.OtherTing, player.SafeBefore, player.Actions, room.CountDeadCard(), name}
	info, err := json.Marshal(gossipInfo)
	if err != nil {
		fmt.Println(err)
	}

	game.GossipDealer.Emit("gossipInfo", string(info), func(pid int, sentence string) {
		fmt.Println(pid, sentence)
		room.BroadcastCoversation(pid, sentence)
	})
	// room.BroadcastCoversation(player.ID, action)
}
