package service

import (
	"encoding/json"
	"fmt"
	"speedtest-gd/global"
	"speedtest-gd/runtimes"
	"speedtest-gd/utils"
)

func initNode() string {
	err := runtimes.InitGlobal()
	info := ""
	if err != nil {
		info += fmt.Sprintf("[x] 在线配置加载失败: %v\n", err)
	}
	LocalPath := "local"
	localAgents, err := runtimes.LoadAllLocalApacheAgents(LocalPath)
	if err != nil {
		info += fmt.Sprintf("[x] 本地配置加载失败: %v\n", err)
	} else {
		global.GlobalApacheAgents = utils.MergeUnique(global.GlobalApacheAgents, localAgents)
	}
	bestNode, err := runtimes.SelectBestNode()
	if err != nil {
		info += fmt.Sprintf("[x] 获取最佳节点失败: %v", err)
		if len(global.GlobalApacheAgents) > 0 {
			global.GlobalBestAgent = global.GlobalApacheAgents[0]
		}
	} else {
		global.GlobalBestAgent = *bestNode
	}
	if len(global.GlobalApacheAgents) < 1 {
		return info
	} else {
		return "OK"
	}
}

func getInfo() string {
	jsonBytes, err := json.Marshal(global.GlobalClientInfo)
	if err != nil {
		return "1"
	}
	return string(jsonBytes)
}
