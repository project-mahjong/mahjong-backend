package core

type Mahjong struct {
}

func NewMahjong() *Mahjong {
	return &Mahjong{}
}

func (m *Mahjong) Start(request *StartRequest) (response *Response, err error) {
	return nil, nil
}

func (m *Mahjong) Next(request *Request) (response *Response, err error) {
	return nil, nil
}
