package core

type Mahjong struct {
}

func NewMahjong() *Mahjong {
	return &Mahjong{}
}

func (m *Mahjong) Start(request string) (response string, err error) {
	return "", nil
}

func (m *Mahjong) Next(request string) (response string, err error) {
	return "", nil
}
