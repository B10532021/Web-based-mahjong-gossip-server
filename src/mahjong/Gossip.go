package mahjong

import (
	"fmt"
	"math/rand"
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

type Record struct {
	Times      int
	Continuous int
}

func (record Record) Clear() {
	record.Times = 0
	record.Continuous = 0
}

func (room Room) Speak(action string, card Tile, currentIdx int, actionIdx int) {
	//room.BroadcastCoversation(currentIdx, action)
	Ba := false
	var BaAction string
	if action == "Throw" {
		room.BroadcastCoversation(actionIdx, card.ToString()) //說打出的牌
		if room.Players[currentIdx].OtherTing || room.Deck.Count() <= 30 {
			if !room.CheckIsSafe(card) {
				action = "Dangerous"
			} else {
				if Find(card, room.Players[currentIdx].SafeBefore) {
					action = "Follow"
				}
				room.Players[currentIdx].SafeBefore = room.Players[currentIdx].SafeBefore[:0]
			}

			for i := 1; i < 3; i++ {
				if otherPlayer := room.Players[(currentIdx+4-i)%4]; otherPlayer.CheckEat(card) {
					otherPlayer.Hand.Add(card)
					if otherPlayer.StepsToHu > otherPlayer.Hand.CountStepsToHu() {
						room.BroadcastCoversation(otherPlayer.ID, "WantEat")
					}
					otherPlayer.Hand.Sub(card)
				}
			}

		} else if room.Players[currentIdx].ThrowTimes < 1 && card.Suit != 3 { //別人前兩輪不是丟字牌
			action = "ThrowGoodFirst"
			actionIdx = (currentIdx + rand.Intn(3) + 1) % 4
		}
	} else if action == "Draw" {
		if card.Suit == 4 {
			if flowers := room.Players[currentIdx].Flowers.Count(); flowers >= 6 {
				if flowers == room.CountDeadCard(card) {
					action = "LotsOfFlowers"
				}
			}
		} else if room.Players[currentIdx].StepsToHu > room.Players[currentIdx].Hand.CountStepsToHu() {
			action = "SubTingNum"
		} else if Find(card, room.Players[currentIdx].ThrowBefore) {
			action = "ThrowBefore"
		} else if room.Players[currentIdx].OtherTing || room.Deck.Count() <= 48 {
			if room.CheckIsUseless(card) {
				if room.Players[currentIdx].IsTing || room.Players[currentIdx].StepsToHu <= 2 {
					action = "Useless"
				} else {
					action = "Safe"
				}
			}
		}
	} else if action == "Ting" {
		for i := 1; i < 4; i++ {
			room.Players[(currentIdx+1)%4].OtherTing = true
		}
		if room.Deck.Count() > 60 {
			action += "Fast"
		}
	}

	if action == "Pon" || action == "Eat" || action == "Gon" || action == "Hu" {
		Ba = true
		BaAction = "Ba" + action
	}

	if action == "Pon" || action == "Eat" || action == "CantEat" || action == "Gon" {
		if action == "Pon" && room.Players[actionIdx].PonTiles.Count() > 3 && room.Players[actionIdx].Times[0].Times == 0 {
			action += "FourNoEat"
		} else if CountTimes(room.Players[actionIdx].Times) >= 4 {
			action += "ManyTimes"
		} else if room.CheckLastNeed(card, room.Players[actionIdx].UUID) && action != "Gon" {
			action += "LastCard"
		} else if room.Players[actionIdx].Times[Condition[action]].Continuous >= 2 {
			action += "TwoSeq"
		} else if room.Players[actionIdx].Times[Condition[action]].Times >= 3 {
			action += "ThreeUp"
		}
	}

	if action != "Draw" && action != "Throw" {
		room.BroadcastCoversation(actionIdx, action)
	}

	if Ba {
		if BaAction != "BaHu" {
			if CountTimes(room.Players[currentIdx].BaTimes) >= 5 {
				BaAction = "BaManyTimes"
			} else if room.Players[currentIdx].BaTimes[Condition[action]].Continuous >= 2 {
				BaAction += "TwoSeq"
			} else if room.Players[currentIdx].BaTimes[Condition[action]].Times >= 3 {
				BaAction += "ThreeUp"
			}
		}
		room.BroadcastCoversation(currentIdx, BaAction)

		if sub := (actionIdx - currentIdx + 4) % 4; sub > 1 && BaAction != "BaHu" { //打牌的下家被跳過
			room.BroadcastCoversation((currentIdx+1+rand.Intn(sub-1))%4, "PassDraw")
		}

	} else if action == "Dangerous" || action == "Follow" || action == "Ting" || action == "TingFast" || action == "Ongon" || action == "KeepWin" {
		action = "Other" + action
		room.BroadcastCoversation((actionIdx+rand.Intn(3)+1)%4, action)
	}

}

func (player *Player) Action(Command int, throwIdx int) {
	switch Command {
	case COMMAND["NONE"]:
		player.Times[1].Continuous = 0
		if throwIdx == (player.ID+3)%4 {
			player.Times[0].Continuous = 0
		} else {
			player.Times[2].Continuous = 0
		}

	case COMMAND["EAT"]:
		player.Times[0].Times++
		player.Times[0].Continuous++
		player.Times[1].Continuous = 0
		player.Times[2].Continuous = 0
	case COMMAND["PON"]:
		player.Times[1].Times++
		player.Times[0].Continuous = 0
		player.Times[1].Continuous++
		player.Times[2].Continuous = 0
	case COMMAND["GON"]:
		player.Times[2].Times++
		player.Times[0].Continuous = 0
		player.Times[1].Continuous = 0
		player.Times[2].Continuous++
	default:
		fmt.Println("Wrong Command")
	}
}

func (player *Player) BaAction(Command int) {
	switch Command {
	case COMMAND["NONE"]:
		for i := 0; i < 3; i++ {
			player.BaTimes[i].Continuous = 0
		}
	case COMMAND["PON"]:
		player.BaTimes[0].Continuous = 0
		player.BaTimes[1].Continuous++
		player.BaTimes[2].Continuous = 0
	case COMMAND["GON"]:
		player.BaTimes[0].Continuous = 0
		player.BaTimes[1].Continuous = 0
		player.BaTimes[2].Continuous++
	case COMMAND["EAT"]:
		player.BaTimes[0].Continuous++
		player.BaTimes[1].Continuous = 0
		player.BaTimes[2].Continuous = 0
	default:
		fmt.Println("Wrong Command")
	}
}

//前面自己丟過的牌
func (player *Player) AddThrowBefore(card Tile) {
	player.ThrowBefore = append(player.ThrowBefore, card)
	if len(player.ThrowBefore) > 3 {
		player.ThrowBefore = player.ThrowBefore[1:]
	}
}

func (room Room) CheckIsUseless(card Tile) bool {
	IsUseless := false
	var amount uint
	amount = room.CountDeadCard(card)
	if amount >= 2 {
		IsUseless = true
	}

	return IsUseless
}

func (room Room) CheckLastNeed(card Tile, selfuid string) bool {
	IsLast := false
	var amount uint
	amount = room.CountDeadCard(card)
	for i := 0; i < 4; i++ {
		if room.Players[i].UUID == selfuid {
			amount += room.Players[i].Hand[card.Suit].GetIndex(card.Value)
			break
		}
	}
	if amount == 4 {
		IsLast = true
	}

	return IsLast
}

func (room Room) CheckIsSafe(card Tile) bool {
	var amount uint
	amount = room.CountDeadCard(card)
	if amount <= 2 {
		return false
	} else if amount >= 3 {
		return true
	}

	return true
}

func (room Room) CheckKeepWin(keepwin bool) bool {
	if keepwin {
		return true
	} else {
		return false
	}
}

func Find(item Tile, array []Tile) bool {
	for _, element := range array {
		if item.Suit == element.Suit && item.Value == element.Value {
			return true
		}
	}
	return false
}

func (room Room) CountDeadCard(card Tile) uint {
	var amount uint = 0

	for i := 0; i < 4; i++ {
		if card.Suit == suitMap["f"] {
			amount += room.Players[i].Flowers.Count()
		} else {
			amount += room.Players[i].DiscardTiles[card.Suit].GetIndex(card.Value)
			amount += room.Players[i].PonTiles[card.Suit].GetIndex(card.Value) * 3
			amount += room.Players[i].GonTiles[card.Suit].GetIndex(card.Value) * 4
			for _, eat := range room.Players[i].EatTiles {
				if eat.First.Suit == card.Suit && card.Value-eat.First.Value < 3 {
					amount++
				}
			}
		}

	}
	return amount
}

func CountTimes(Times [3]Record) int {
	return Times[0].Times + Times[1].Times + Times[2].Times
}
