package service

import (
	"encoding/json"
	"fmt"
	"speedtest-gd/global"
)

func showBestNode() string {
	jsonData, err := json.Marshal(global.GlobalBestAgent)
	if err != nil {
		Error := fmt.Sprintf("%v", err)
		return Error
	} else {
		return string(jsonData)
	}
}
