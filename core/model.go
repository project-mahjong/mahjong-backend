package core

type Tile string
type Group struct {
	Type        int
	From        int
	BaseTile    Tile
	CallingTile Tile
}
type StartRequest struct {
	PrevailingWind int
	LianZhuang     int
}

type Request struct {
	Discard int
	OK      bool
}

type Response struct {
	Error       int
	ErrorString string
	Title       struct {
		Wall                 Tile
		MD5                  string
		DoraIndicatorCount   int
		ReplacementTileCount int
		WallCount            int
		Player               [4]struct {
			HandTile []Tile
			NowTile  Tile
			PaiHe    []Tile
			TingPai  int
			Riichi   int
			Groups   []Group `json:"Group"`
		}
	}
	Action struct {
		Type   int
		Player []struct {
			CanDiscard []bool
			Groups     []Group `json:"Group"`
		}
	}
	End struct {
		Player [4]struct {
			IsWin      int
			Yaku       []int
			Minipoints int
			Score      int
		}
	}
}
