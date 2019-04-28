package mahjong

import (
	"math/rand"
	"strconv"
)

var GroupMask = [40]int32{
	02, 020, 0200, 02000, 020000, 0200000, 02000000, 020000000, 0200000000,
	03, 030, 0300, 03000, 030000, 0300000, 03000000, 030000000, 0300000000,
	0111, 01110, 011100, 0111000, 01110000, 011100000, 0111000000,
	011, 0110, 01100, 011000, 0110000, 01100000, 011000000, 0110000000,
	0101, 01010, 010100, 0101000, 01010000, 010100000, 0101000000,
}

type Groups struct {
	Code     []uint32
	Hu_Tiles []uint32
	Steps    int
}

// SuitSet represents a set of suit
type SuitSet [5]Suit

// SuitTileCount represents each suit's tile count
var SuitTileCount = [5]uint{9, 9, 9, 7, 8}

// NewSuitSet creates a new suit suitSet
func NewSuitSet(full bool) SuitSet {
	var suitSet SuitSet
	t, _ := strconv.ParseUint("100100100100100100100100100", 2, 32)
	for s := 0; s < 3; s++ {
		suitSet[s] = IF(full, Suit(t), Suit(0)).(Suit)
	}

	t, _ = strconv.ParseUint("100100100100100100100", 2, 32) // 字、風
	suitSet[3] = IF(full, Suit(t), Suit(0)).(Suit)

	t, _ = strconv.ParseUint("001001001001001001001001", 2, 32)
	suitSet[4] = IF(full, Suit(t), Suit(0)).(Suit)

	return suitSet
}

// ArrayToSuitSet converts tile array to suit set
func ArrayToSuitSet(tile []Tile) SuitSet {
	suitSet := NewSuitSet(false)
	suitSet.Add(tile)
	return suitSet
}

// IsEmpty returns if suit set is empty
func (suitSet SuitSet) IsEmpty() bool {
	count := uint32(0)
	for i := 0; i < 5; i++ {
		count += uint32(suitSet[i])
	}
	return count == 0
}

// IsContainColor return if suit set contain color
func (suitSet SuitSet) IsContainColor(color int) bool {
	return uint32(suitSet[color]) > 0
}

// Have returns if suit set has tlie
func (suitSet SuitSet) Have(tile Tile) bool {
	return suitSet[tile.Suit].GetIndex(tile.Value) > 0
}

// At returns idx th tile in suit set
func (suitSet SuitSet) At(idx int) Tile {
	amount := 0
	for s := 0; s < 5; s++ {
		for v := uint(0); v < SuitTileCount[s]; v++ {
			amount += int(suitSet[s].GetIndex(v))
			if amount > idx {
				return NewTile(s, v)
			}
		}
	}
	return NewTile(-1, 0)
}

// Count returns amount of suit set
func (suitSet SuitSet) Count() uint {
	amount := uint(0)
	for s := 0; s < 5; s++ {
		amount += suitSet[s].Count()
	}
	return amount
}

// Draw draws a tile from suit set
func (suitSet *SuitSet) Draw() Tile {
	amount := int32(suitSet.Count())
	tile := suitSet.At(int(rand.Int31n(amount)))
	suitSet.Sub(tile)
	return tile
}

// ToStringArray converts suit set to string array
func (suitSet SuitSet) ToStringArray() []string {
	var result []string
	for s := 0; s < 5; s++ {
		for v := uint(0); v < SuitTileCount[s]; v++ {
			for n := uint(0); n < suitSet[s].GetIndex(v); n++ {
				result = append(result, suitStr[s]+strconv.Itoa(int(v+1)))
			}
		}
	}
	return result
}

// Add adds a tile or a suit set to a suit set
func (suitSet *SuitSet) Add(input interface{}) {
	switch input.(type) {
	case []Tile:
		for _, tile := range input.([]Tile) {
			if suitSet[tile.Suit].GetIndex(tile.Value) < 4 {
				suitSet.Add(tile)
			}
		}
	case Tile:
		tile := input.(Tile)
		if suitSet[tile.Suit].GetIndex(tile.Value) < 4 {
			suitSet[tile.Suit].Add(tile.Value)
		}
	}
}

// Sub subs a tile or a suit set from a suit set
func (suitSet *SuitSet) Sub(input interface{}) {
	switch input.(type) {
	case []Tile:
		for _, tile := range input.([]Tile) {
			if suitSet[tile.Suit].GetIndex(tile.Value) > 0 {
				suitSet.Sub(tile)
			}
		}
	case Tile:
		tile := input.(Tile)
		if suitSet[tile.Suit].GetIndex(tile.Value) > 0 {
			suitSet[tile.Suit].Sub(tile.Value)
		}
	}
}

func (suitSet SuitSet) CountStepsToHu() int {
	var ret []int32
	var i int32
	for i = 0; i < 138; i++ {
		if HaveComponent(int32(suitSet[i/40]), GroupMask[i%40]) {
			ret = append(ret, i)
		}
	}

	var result []Groups
	var tmp []uint32
	FindGroups(suitSet, &ret, &result, tmp, 0)

	var G []Groups
	G = CountSteps(&result)
	return G[0].Steps
}

func HaveComponent(suit int32, component int32) bool {
	var i uint
	for i = 0; i < 9; i++ {
		//fmt.Println((suit>>(i*3)&7), ((component>>(i*3))&7))
		if (suit>>(i*3))&7 < (component>>(i*3))&7 {
			return false
		}
	}
	return true
}

func FindGroups(suitSet SuitSet, Components *[]int32, result *[]Groups, tmp []uint32, start int) {
	//fmt.Println(len(*Components))
	f := true
	for i := start; i < len(*Components); i++ {
		if HaveComponent(int32(suitSet[(*Components)[i]/40]), GroupMask[(*Components)[i]%40]) {
			f = false
			tiles := suitSet
			tiles[(*Components)[i]/40] -= Suit(GroupMask[(*Components)[i]%40])
			tmp = append(tmp, uint32((*Components)[i]))
			FindGroups(tiles, Components, result, tmp, i)
			tmp = tmp[:len(tmp)-1]

		}
	}

	if f {
		for i := 0; i < 34; i++ {
			if (suitSet[i/9] >> uint(i%9*3) & 7) > 0 {
				for j := 0; j < int(suitSet[i/9])>>uint(i%9*3)&7; j++ {
					tmp = append(tmp, uint32(180+i))
				}
			}
		}
		var G Groups
		G.Code = tmp
		*result = append(*result, G)
	}
}

func CountSteps(G *[]Groups) []Groups {
	var ret []Groups

	min := 100
	var total, single, pair, group, pre_group, steps int
	for i := 0; i < len(*G); i++ {
		single, pair, group, pre_group = 0, 0, 0, 0
		for j := 0; j < len((*G)[i].Code); j++ {
			if (*G)[i].Code[j] >= 180 {
				single++
			} else if (*G)[i].Code[j]%40 < 9 {
				pair++
			} else if (*G)[i].Code[j]%40 < 25 {
				group++
			} else {
				pre_group++
			}
		}
		total = group*3 + pre_group*2 + pair*2 + single
		steps = total/3*2 + 1
		steps -= group*2 + pre_group + pair

		if group+pre_group+pair > total/3+1 {
			steps += group + pre_group + pair - (total/3 + 1)
		}

		if pre_group != 0 && pre_group >= (total/3+1-group) && pair&^pair != 0 {
			steps++
		}

		if steps < min {
			min = steps
		}

		(*G)[i].Steps = steps
	}

	for i := 0; i < len(*G); i++ {
		if (*G)[i].Steps <= min+0 {
			ret = append(ret, (*G)[i])
		}
	}

	return ret
}
