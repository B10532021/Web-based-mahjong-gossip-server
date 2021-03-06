package mahjong

import (
	"log"
	"math/rand"
	"sync"
	"time"

	socketio "github.com/googollee/go-socket.io"
	uuid "github.com/satori/go.uuid"
)

const (
	MONGODB_URL = "127.0.0.1:27017"
)

var game *GameManager

// NewGameManager creates a new gameManager
func NewGameManager() bool {
	server, err := socketio.NewServer([]string{"websocket"})
	if err != nil {
		log.Fatal(err)
		return true
	}
	// session, err2 := mgo.Dial(MONGODB_URL)
	// if err2 != nil {
	// 	panic(err2)
	// }
	// session.SetMode(mgo.Monotonic, true)
	// db := session.DB("gossip")

	rooms := make(map[string]*Room)
	game = &GameManager{rooms, server, sync.Mutex{}, nil}
	return false
}

// GameManager represents a gameManager
type GameManager struct {
	Rooms  map[string]*Room
	Server *socketio.Server
	Locker sync.Mutex
	//GossipDB     *mgo.Database
	GossipDealer socketio.Socket
}

// GetServer returns socket io server
func GetServer() *socketio.Server {
	return game.Server
}

// Login handles player's login
func Login(name string, socket *socketio.Socket) (string, bool) {
	uuid, err := AddPlayer(name)
	if err {
		return "", true
	}
	index := FindPlayerByUUID(uuid)
	PlayerList[index].Socket = socket
	PlayerList[index].State = WAITING

	return uuid, false
}

// Logout handles player's logout
func Logout(index int) {
	game.Locker.Lock()
	if index >= 0 && index < len(PlayerList) {
		if PlayerList[index].State == WAITING {
			RemovePlayer(index)
		} else if PlayerList[index].State == MATCHED && PlayerList[index].LeaveCount == 0 {
			RemoveRoom(PlayerList[index].Room)
			RemovePlayer(index)
		} else if PlayerList[index].State == PLAYING {
			players := FindPlayerListInRoom(PlayerList[index].Room, 0)
			flag := true
			for _, player := range players {
				if player.LeaveCount > 0 {
					flag = false
					break
				}
			}
			if flag {
				RemoveRoom(PlayerList[index].Room)
				RemovePlayer(index)
			}
		}
		// else if PlayerList[index].State == MATCHED {
		// 	game.RemoveRoom(PlayerList[index].Room)
		// 	RemovePlayer(index)
		// }
	}
	game.Locker.Unlock()
}

// Exec executes the whole game
func Exec() {
	for {
		if WaitingNum(0) >= 1 && WaitingNum(1) >= 3 {
			go CreateRoom()
			time.Sleep(2 * time.Second)
		}
		time.Sleep(5 * time.Second)
	}
}

// WaitingNum returns the number of player which state are waiting
func WaitingNum(isBot int) int {
	return len(FindPlayerListIsSameState(WAITING, isBot))
}

// CreateRoom creates a new room and add player to that room
func CreateRoom() {
	var roomName string
	for {
		roomName = uuid.Must(uuid.NewV4()).String()
		if game.Rooms[roomName] == nil {
			break
		}
	}
	matchPlayer := Match()
	game.Rooms[roomName] = NewRoom(roomName)
	game.Rooms[roomName].IO = game.Server
	game.Rooms[roomName].AddPlayer(matchPlayer)
	game.Rooms[roomName].WaitToStart()
	RemoveRoom(roomName)
}

// RemoveRoom removes a room by room name
func RemoveRoom(name string) {
	if game.Rooms[name] == nil {
		return
	}
	if game.Rooms[name].Waiting {
		game.Rooms[name].StopWaiting()
	}
	playerList := FindPlayerListInRoom(name, -1)
	for _, player := range playerList {
		var index int
		index = FindPlayerByUUID(player.UUID)
		if game.Rooms[name].Waiting {
			PlayerList[index].State = WAITING
		} else {
			if PlayerList[index].IsBot {
				PlayerList[index].State = WAITING
			} else {
				RemovePlayer(index)
			}
		}
	}
	(*game.Rooms[name]) = Room{}
	delete(game.Rooms, name)
}

// Match matchs 4 player into a room
func Match() []string {
	var sample []string

	waitingBotList := FindPlayerListIsSameState(WAITING, 1)
	for i := 0; i < 3; i++ {
		index := rand.Int31n(int32(len(waitingBotList)))
		sample = append(sample, waitingBotList[index].UUID)
		waitingBotList = append(waitingBotList[:index], waitingBotList[index+1:]...)
	}

	waitingPlayerList := FindPlayerListIsSameState(WAITING, 0)
	index := rand.Int31n(int32(len(waitingPlayerList)))
	sample = append(sample, waitingPlayerList[index].UUID)

	// waitingPlayerList := FindPlayerListIsSameState(WAITING, 0)
	// for i := 0; i < 4; i++ {
	// 	index := rand.Int31n(int32(len(waitingPlayerList)))
	// 	sample = append(sample, waitingPlayerList[index].UUID)
	// 	waitingPlayerList = append(waitingPlayerList[:index], waitingPlayerList[index+1:]...)
	// }

	for _, uuid := range sample {
		index := FindPlayerByUUID(uuid)
		PlayerList[index].State = MATCHED
	}
	return sample
}
