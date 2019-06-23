package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/WAAutoMaton/mahjong-backend/core"
	"log"
	"os"
)

func main() {
	m := core.NewMahjong()
	cin := bufio.NewReader(os.Stdin)
	requestString, _, err := cin.ReadLine()
	requestString = []byte(`{"PrevailingWind":0,"RemainingDealer":0,"Riichi":[true,true,true,true]}`)
	err = nil
	if err != nil {
		log.Panic("unable to read stdin")
	}
	request := &core.StartRequest{}
	err = json.Unmarshal(requestString, request)
	if err != nil {
		fmt.Println(core.NewJsonError(""))
		return
	}
	response, err := m.Start(request)
	if err != nil {
		fmt.Println(err)
		return
	}
	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println(core.NewUnknownError("Unexpected error while marshaling json."))
		return
	}
	fmt.Println(string(data))
	for {
		requestString, _, err := cin.ReadLine()
		if err != nil || string(requestString) == "" {
			requestString = []byte(`[{"Discard":5}]`)
			err = nil
		}
		if err != nil {
			log.Panic("unable to read stdin")
		}
		var request []core.Request
		err = json.Unmarshal(requestString, &request)
		if err != nil {
			fmt.Println(core.NewJsonError(""))
			return
		}
		response, err := m.Next(request)
		if err != nil {
			fmt.Println(core.NewUnknownError("Unknown error in main progress."))
			return
		}
		data, err := json.Marshal(response)
		if err != nil {
			fmt.Println(core.NewUnknownError("Unexpected error while marshaling json."))
			return
		}
		fmt.Println(string(data))
		if _, ok := response.(*core.ResponseEnd); ok {
			return
		}
	}
}
