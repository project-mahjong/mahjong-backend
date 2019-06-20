package core

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"strconv"
)

type Mahjong struct {
	prevailingWind   int
	remainingDealer  int
	wall             [136]Tile
	wallCount        int
	doraCount        int
	ReplacementCount int
	md5              string
	player           [4]Player
}

type Player struct {
	handTile       []Tile
	nowTile        Tile
	discardTile    []Tile
	discardTo      []int // 表示舍牌到哪里去了
	scoreCanRiichi bool
	riichiTile     int // 立直宣言牌为舍牌的第几张,未立直则为-1
	group          []Group
}

func MakeError(id int, errorString string) string {
	return fmt.Sprintf(`{"Error":%d,"ErrorString":%s}`, id, errorString)
}

func NewMahjong() *Mahjong {
	m := &Mahjong{}
	m.initWall()
	for i := 0; i < 4; i++ {
		m.player[i].init()
	}
	m.takeTile()
	return m
}

type UnknownError struct {
	errorString string
}

func (e *UnknownError) Error() string {
	return MakeError(-1, e.errorString)
}

type JsonError struct {
	errorString string
}

func (e *JsonError) Error() string {
	return MakeError(-2, e.errorString)
}

type InvalidValueError struct {
	errorString string
}

func (e *InvalidValueError) Error() string {
	return MakeError(-3, e.errorString)
}

func (m *Mahjong) Start(request *StartRequest) (response *ResponseAction, err error) {
	if request.PrevailingWind < 0 || request.PrevailingWind >= 4 {
		return nil, &InvalidValueError{"PrevailingWind invalid"}
	}

	m.prevailingWind = request.PrevailingWind
	m.remainingDealer = request.RemainingDealer
	for i := 0; i < 4; i++ {
		m.player[i].scoreCanRiichi = request.Riichi[i]
	}
	response = &ResponseAction{Response: *m.getTitle()}
	response.Error = 0
	response.ErrorString = ""
	response.Action.Type = 0
	canDiscard := make([]bool, 14)
	for i := 0; i < 14; i++ {
		canDiscard[i] = true
	}
	response.Action.Player = make([]ResponseActionPlayer, 1)
	response.Action.Player[0].CanDiscard = canDiscard
	response.Action.Player[0].ID = 0
	return response, nil
}

func (m *Mahjong) Next(request *Request) (response *interface{}, err error) {
	return nil, nil
}

func (m *Mahjong) initWall() {
	cnt := 0
	for i := 1; i <= 7; i++ {
		for j := 1; j <= 4; j++ {
			m.wall[cnt] = Tile(strconv.Itoa(i) + "z")
			cnt++
		}
	}
	for i := 1; i <= 9; i++ {
		for j := 1; j <= 4; j++ {
			t := i
			if i == 5 && j == 4 {
				t = 0
			}
			m.wall[cnt] = Tile(strconv.Itoa(t) + "m")
			cnt++
			m.wall[cnt] = Tile(strconv.Itoa(t) + "s")
			cnt++
			m.wall[cnt] = Tile(strconv.Itoa(t) + "p")
			cnt++
		}
	}
	for i := len(m.wall) - 1; i > 0; i-- {
		t := rand.Intn(i + 1)
		m.wall[i], m.wall[t] = m.wall[t], m.wall[i]
	}
	data := ""
	for i := 53; i < 136; i++ {
		data += string(m.wall[i])
	}
	m.md5 = fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

func (m *Mahjong) takeTile() {
	cnt := 0
	for i := 0; i < 4; i++ {
		for j := 1; j <= 13; j++ {
			m.player[i].handTile = append(m.player[i].handTile, m.wall[cnt])
			cnt++
		}
	}
	m.player[0].handTile = append(m.player[0].handTile, m.wall[cnt])
	m.wallCount = 53
	m.doraCount = 1
	m.ReplacementCount = 0
}

func (m *Mahjong) getTitle() (response *Response) {
	response = &Response{}
	response.Title.Wall = m.wall
	response.Title.MD5 = m.md5
	response.Title.WallCount = m.wallCount
	response.Title.DoraIndicatorCount = m.doraCount
	response.Title.ReplacementTileCount = m.ReplacementCount
	for i := 0; i < 4; i++ {
		response.Title.Player[i].HandTile = m.player[i].handTile
		response.Title.Player[i].NowTile = m.player[i].nowTile
		river := make([]Tile, 0)
		for j := 0; j < len(m.player[i].discardTile); j++ {
			if m.player[i].discardTo[j] == i {
				river = append(river, m.player[i].discardTile[j])
			}
		}
		response.Title.Player[i].DiscardTile = river
		response.Title.Player[i].ReadHand = m.player[i].isReadHand()
		if m.player[i].riichiTile == -1 {
			response.Title.Player[i].Riichi = -1
		} else {
			t := m.player[i].riichiTile
			for m.player[i].discardTo[t] != i && t < len(m.player[i].discardTo) {
				t++
			}
			if t == len(m.player[i].discardTo) {
				response.Title.Player[i].Riichi = -2
			} else {
				cnt := 0
				for j := 0; j <= t; j++ {
					if m.player[i].discardTo[j] == i {
						cnt++
					}
				}
				response.Title.Player[i].Riichi = cnt
			}
		}
		response.Title.Player[i].Groups = m.player[i].group
	}
	return response
}

func (p *Player) init() {
	p.discardTile = make([]Tile, 0)
	p.discardTo = make([]int, 0)
	p.riichiTile = -1
	p.group = make([]Group, 0)
}

// 该玩家是否听牌
// 返回值：　0:否  1:是  2:振听
func (p *Player) isReadHand() int {
	//TODO: 听牌判定
	return 0
}
