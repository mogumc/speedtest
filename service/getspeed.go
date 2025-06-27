package service

import (
	"encoding/json"
	"fmt"
	"speedtest-gd/global"
)

func getspeed() string {
	jsonData, err := json.Marshal(global.GlobalSpeed)
	if err != nil {
		Error := fmt.Sprintf("%v", err)
		return Error
	} else {
		return string(jsonData)
	}
}
