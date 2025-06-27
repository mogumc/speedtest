package service

import (
	"encoding/json"
	"fmt"
	"speedtest-gd/global"
	"speedtest-gd/runtimes"
)

func showBestNode() string {
	bestNode, err := runtimes.SelectBestNode()
	global.GlobalBestAgent = *bestNode
	if err != nil {
		Error := fmt.Sprintf("%v", err)
		return Error
	}
	jsonData, err := json.Marshal(bestNode)
	if err != nil {
		Error := fmt.Sprintf("%v", err)
		return Error
	} else {
		return string(jsonData)
	}
}
