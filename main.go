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
	if err != nil {
		log.Panic("unable to read stdin")
	}
	request := &core.StartRequest{}
	err = json.Unmarshal(requestString, request)
	if err != nil {
		fmt.Println(`{"Error":-2,"ErrorString":""}`)
		return
	}
	response, err := m.Start(request)
	if err != nil {
		fmt.Println(`{"Error":-1,"ErrorString":"Unknown error."}`)
		return
	}
	fmt.Println(json.Marshal(response))
	for {
		requestString, _, err := cin.ReadLine()
		if err != nil {
			log.Panic("unable to read stdin")
		}
		request := &core.Request{}
		err = json.Unmarshal(requestString, request)
		if err != nil {
			fmt.Println(`{"Error":-2,"ErrorString":""}`)
			return
		}
		response, err := m.Next(request)
		if err != nil {
			fmt.Println(`{"Error":-1,"ErrorString":"Unknown error."}`)
			return
		}
		fmt.Println(json.Marshal(response))
	}
}
