package core

import (
	"crypto/md5"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strconv"
)

var ID2Tile [136]Tile

func init() {
	cnt := 0
	for k := 1; k <= 3; k++ {
		for i := 1; i <= 9; i++ {
			for j := 1; j <= 4; j++ {
				t := i
				if i == 5 && j == 1 {
					t = 0
				}
				if k == 1 {
					ID2Tile[cnt] = Tile(strconv.Itoa(t) + "m")
				} else if k == 2 {
					ID2Tile[cnt] = Tile(strconv.Itoa(t) + "p")
				} else if k == 3 {
					ID2Tile[cnt] = Tile(strconv.Itoa(t) + "s")
				}
				cnt++
			}
		}
	}
	for i := 1; i <= 7; i++ {
		for j := 1; j <= 4; j++ {
			ID2Tile[cnt] = Tile(strconv.Itoa(i) + "z")
			cnt++
		}
	}
}

type Mahjong struct {
	prevailingWind   int
	remainingDealer  int
	wall             [136]int
	wallCount        int
	doraCount        int
	ReplacementCount int
	md5              string
	player           [4]Player
	turnTo           int // 当前应谁出牌
	lastActionType   int
	lastTile         int
	winList          []int
}

type Player struct {
	handTile       []int
	discardTile    []int
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

func NewUnknownError(errorString string) error {
	return &UnknownError{errorString}
}

func (e *UnknownError) Error() string {
	return MakeError(-1, e.errorString)
}

type JsonError struct {
	errorString string
}

func NewJsonError(errorString string) error {
	return &JsonError{errorString}
}

func (e *JsonError) Error() string {
	return MakeError(-2, e.errorString)
}

type InvalidValueError struct {
	errorString string
}

func NewInvalidValueError(errorString string) error {
	return &InvalidValueError{errorString}
}

func (e *InvalidValueError) Error() string {
	return MakeError(-3, e.errorString)
}

func (m *Mahjong) Start(request *StartRequest) (response *ResponseAction, err error) {
	if request.PrevailingWind < 0 || request.PrevailingWind >= 4 {
		return nil, NewInvalidValueError("PrevailingWind invalid")
	}
	if request.RemainingDealer < 0 || request.RemainingDealer >= 4 {
		return nil, &InvalidValueError{"RemainingDealer invalid"}
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

func (m *Mahjong) Next(request []Request) (response interface{}, err error) {
	nowPlayer := &m.player[m.turnTo]
	if m.lastActionType == 0 {
		if len(request) != 1 {
			return nil, &InvalidValueError{"Array size not equal to 1"}
		}
		discard := request[0].Discard
		if discard < 0 || discard >= len(m.player[m.turnTo].handTile) {
			return nil, &InvalidValueError{"Discard invalid"}
		}
		tile := nowPlayer.handTile[discard]
		m.lastTile = tile
		removeInt(&nowPlayer.handTile, discard)
		appendInt(&nowPlayer.discardTile, tile)
		appendInt(&nowPlayer.discardTo, m.turnTo)

		m.winList = make([]int, 0)
		for i := 0; i < 4; i++ {
			if i == m.turnTo {
				continue
			}
			t := make([]int, len(m.player[i].handTile)+1)
			copy(t, m.player[i].handTile)
			appendInt(&t, tile)
			if ok, _ := checkWin(t, m.player[i].group); ok {
				appendInt(&m.winList, i)
			}
		}
		if len(m.winList) > 0 {
			m.lastActionType = 2
			res := &ResponseAction{Response: *m.getTitle()}
			res.Action.Type = 2
			for _, i := range m.winList {
				res.Action.Player = append(res.Action.Player, ResponseActionPlayer{ID: i, CanDiscard: nil, Groups: nil})
			}
			return res, nil
		}
		// 进入常规处理部分
	} else if m.lastActionType == 1 {
		//TODO: 吃碰杠处理
		return response, nil
	} else if m.lastActionType == 2 {
		if len(request) != len(m.winList) {
			return nil, &InvalidValueError{}
		}
		winner := make([]int, 0)
		for i := range request {
			if request[i].OK {
				appendInt(&winner, m.winList[i])
			}
		}
		if len(winner) != 0 {
			res := &ResponseEnd{Response: *m.getTitle()}
			for i := 0; i < 4; i++ {
				win := false
				for _, j := range winner {
					if i == j {
						win = true
						break
					}
				}
				if win {
					res.End.Player[i].IsWin = 1
					t := make([]int, len(m.player[i].handTile)+1)
					copy(t, m.player[i].handTile)
					appendInt(&t, m.lastTile)
					_, res.End.Player[i].Yaku = checkWin(t, m.player[i].group)
					res.End.Player[i].Minipoints = 2 //TODO: 计算符数
					res.End.Player[i].Score = 1      //TODO: 计算点数
				} else {
					res.End.Player[i].IsWin = 0
					if i == m.turnTo { //放铳选手
						res.End.Player[i].Score = -1 //TODO: 计算点数
					} else {
						res.End.Player[i].Score = 0
					}
				}
			}
			return res, nil
		}
		// 没人选择和牌就进入常规处理部分
	} else if m.lastActionType == 3 {
		//TODO: 九种九牌处理
		return response, nil
	} else if m.lastActionType == 4 {
		//TODO: 立直处理
		return response, nil
	} else if m.lastActionType == 5 {
		if len(request) != 1 {
			return nil, &InvalidValueError{"array size not equal to 1"}
		}
		if request[0].OK {
			// 和了，终局！
			res := &ResponseEnd{Response: *m.getTitle()}
			_, yaku := checkWin(m.player[m.turnTo].handTile, m.player[m.turnTo].group)
			res.End.Player[m.turnTo].Yaku = yaku
			minipoints := 5 //TODO: 符数计算
			score := 10     //TODO: 得点计算
			if m.turnTo == 0 {
				score += score / 2
			}
			res.End.Player[m.turnTo].Minipoints = minipoints
			res.End.Player[m.turnTo].Score = score
			res.End.Player[m.turnTo].IsWin = 2
			if m.turnTo == 0 {
				score /= 3
			} else {
				score /= 4
			}
			for i := 0; i < 4; i++ {
				if i != m.turnTo {
					res.End.Player[i].IsWin = 0
					if i == 0 {
						res.End.Player[i].Score = -score * 2
					} else {
						res.End.Player[i].Score = -score
					}
					res.End.Player[i].Minipoints = 0
					res.End.Player[i].Yaku = nil
				}
			}
			return res, nil
		} else {
			// 居然不和
			res := &ResponseAction{Response: *m.getTitle()}
			res.Error = 0
			res.ErrorString = ""
			m.lastActionType = 0
			res.Action.Type = 0
			canDiscard := make([]bool, len(nowPlayer.handTile))
			for i := range canDiscard {
				canDiscard[i] = true
			}
			res.Action.Player = make([]ResponseActionPlayer, 1)
			res.Action.Player[0].CanDiscard = canDiscard
			res.Action.Player[0].ID = m.turnTo
			return res, nil
		}
	} else {
		log.Panic("lastActionType invalid")
		return nil, &UnknownError{}
	}

	handTile := make([]int, len(nowPlayer.handTile)+1)
	copy(handTile, nowPlayer.handTile)
	appendInt(&handTile, m.lastTile)
	if ok, _ := checkWin(handTile, nowPlayer.group); ok {
		// 可以自摸
		res := &ResponseAction{Response: *m.getTitle()}
		res.Error = 0
		res.ErrorString = ""
		m.lastActionType = 5
		res.Action.Type = 5
		res.Action.Player = make([]ResponseActionPlayer, 1)
		res.Action.Player[0].ID = m.turnTo
		return res, nil
	}

	if m.wallCount >= 136 { //荒牌流局
		res := &ResponseEnd{Response: *m.getTitle()}
		cnt := 0
		for i := 0; i < 4; i++ {
			if m.player[i].isReadHand() != 0 {
				cnt++
			}
		}
		for i := 0; i < 4; i++ {
			res.End.Player[i].IsWin = m.player[i].isReadHand()
			if m.player[i].isReadHand() == 0 {
				switch cnt {
				case 0:
					res.End.Player[i].Score = 0
				case 1:
					res.End.Player[i].Score = -1000
				case 2:
					res.End.Player[i].Score = -1500
				case 3:
					res.End.Player[i].Score = -3000
				}
			} else {
				switch cnt {
				case 1:
					res.End.Player[i].Score = 3000
				case 2:
					res.End.Player[i].Score = 1500
				case 3:
					res.End.Player[i].Score = 1000
				case 4:
					res.End.Player[i].Score = 0
				}
			}
		}
		return res, nil
	}

	m.turnTo++
	if m.turnTo >= 4 {
		m.turnTo = 0
	}
	nowPlayer = &m.player[m.turnTo]
	appendInt(&nowPlayer.handTile, m.wall[m.wallCount])
	m.wallCount++
	res := &ResponseAction{Response: *m.getTitle()}
	res.Error = 0
	res.ErrorString = ""
	res.Action.Type = 0
	canDiscard := make([]bool, len(nowPlayer.handTile))
	for i := range canDiscard {
		canDiscard[i] = true
	}
	res.Action.Player = make([]ResponseActionPlayer, 1)
	res.Action.Player[0].CanDiscard = canDiscard
	res.Action.Player[0].ID = m.turnTo
	response = res
	return response, nil
}

func (m *Mahjong) initWall() {
	copy(m.wall[:], rand.Perm(136))
	data := ""
	for i := 53; i < 136; i++ {
		data += string(ID2Tile[m.wall[i]])
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
	for i := range m.wall {
		response.Title.Wall[i] = ID2Tile[m.wall[i]]
	}
	response.Title.MD5 = m.md5
	response.Title.WallCount = m.wallCount
	response.Title.DoraIndicatorCount = m.doraCount
	response.Title.ReplacementTileCount = m.ReplacementCount
	for i := 0; i < 4; i++ {
		for _, v := range m.player[i].handTile {
			response.Title.Player[i].HandTile = append(response.Title.Player[i].HandTile, ID2Tile[v])
		}
		response.Title.Player[i].NowTile = Tile("")
		river := make([]Tile, 0)
		for j := 0; j < len(m.player[i].discardTile); j++ {
			if m.player[i].discardTo[j] == i {
				river = append(river, ID2Tile[m.player[i].discardTile[j]])
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
		for _, v := range m.player[i].group {
			t := GroupResponse{}
			t.Type = v.Type
			t.CallingTile = ID2Tile[v.CallingTile]
			for _, v2 := range v.Tiles {
				t.Tiles = append(t.Tiles, ID2Tile[v2])
			}
			response.Title.Player[i].Groups = append(response.Title.Player[i].Groups, t)
		}
	}
	return response
}

func (p *Player) init() {
	p.discardTile = make([]int, 0)
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

// 检查玩家是否可以和牌
// ok: 是否能和牌
// yaku: 满足的役种表（不包括宝牌，赤宝牌，里宝牌），有和牌型但无役为空slice，不能和牌则为nil
func checkWin(handTile []int, group []Group) (ok bool, yaku []int) {
	sort.Ints(handTile)
	yaku = make([]int, 0)
	if checkSevenPairs(handTile, group) {
		appendInt(&yaku, 200)
	} else if checkKokushimusou(handTile, group) {
		appendInt(&yaku, 800)
	} else if !checkNormalWin(handTile, 4-len(group), 1) {
		// 没和牌牌型，告辞
		return false, nil
	}
	//TODO: 各种役种检查
	return true, yaku
}

// 七对子型检查
func checkSevenPairs(handTile []int, group []Group) bool {
	if len(group) > 0 {
		return false
	}
	tong := make(map[int]int)
	for _, i := range handTile {
		tong[i/4]++
	}
	for _, i := range tong {
		if i != 2 {
			return false
		}
	}
	return true
}

// 国士无双型检查
func checkKokushimusou(handTile []int, group []Group) bool {
	if len(group) > 0 {
		return false
	}
	flag := make([]bool, 13)
	for _, i := range handTile {
		switch i / 4 {
		case 0:
			flag[0] = true
		case 8:
			flag[1] = true
		case 9:
			flag[2] = true
		case 17:
			flag[3] = true
		case 18:
			flag[4] = true
		case 26:
			flag[5] = true
		default:
			if i/4 >= 27 {
				flag[i/4-21] = true
			} else {
				return false
			}
		}
	}
	for _, i := range flag {
		if !i {
			return false
		}
	}
	return true
}

// 普通和牌型检查（递归函数）
// m: 还需要找出几组面子
func checkNormalWin(handTile []int, m int, q int) bool {
	if len(handTile) != q*2+m*3 {
		log.Panic("handTile size unexpected")
	}
	if m == 0 && q == 0 {
		return true
	}
	if m > 0 && handTile[0]/4 == handTile[1]/4 && handTile[0]/4 == handTile[2]/4 && checkNormalWin(handTile[3:], m-1, q) {
		return true
	}
	if q == 1 && handTile[0]/4 == handTile[1]/4 && checkNormalWin(handTile[2:], m, 0) {
		return true
	}
	if m == 0 {
		return false
	}
	now := handTile[0] / 4
	if now >= 25 || now == 7 || now == 8 || now == 16 || now == 17 {
		return false
	}
	for i := 1; i < len(handTile); i++ {
		if handTile[i]/4 == now+1 {
			for j := i + 1; j < len(handTile); j++ {
				if handTile[j]/4 == now+2 {
					t := make([]int, 0)
					for k := 1; k < len(handTile); k++ {
						if k != i && k != j {
							appendInt(&t, handTile[k])
						}
					}
					return checkNormalWin(t, m-1, q)
				}
			}
			return false
		}
	}
	return false
}
