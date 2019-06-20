package core

type Tile string
type Group struct {
	Type        int
	From        int
	BaseTile    Tile
	CallingTile Tile
}
type StartRequest struct {
	PrevailingWind  int
	RemainingDealer int
	Riichi          [4]bool
}

type Request struct {
	Discard int
	OK      bool
}

type Response struct {
	Error       int
	ErrorString string
	Title       struct {
		Wall                 [136]Tile
		MD5                  string
		DoraIndicatorCount   int
		ReplacementTileCount int
		WallCount            int
		Player               [4]struct {
			HandTile    []Tile
			NowTile     Tile
			DiscardTile []Tile
			ReadHand    int
			Riichi      int
			Groups      []Group `json:"Group"`
		}
	}
}

type ResponseActionPlayer struct {
	ID         int
	CanDiscard []bool
	Groups     []Group `json:"Group"`
}

type ResponseAction struct {
	Response
	Action struct {
		Type   int
		Player []ResponseActionPlayer
	}
}

type ResponseEnd struct {
	Response
	End struct {
		Player [4]struct {
			IsWin      int
			Yaku       []int
			Minipoints int
			Score      int
		}
	}
}
