package service

import (
	"fmt"
	"speedtest-gd/global"
	"speedtest-gd/utils"
)

func pingSelectedNode(id int) string {
	var nodelist *global.ApacheAgent
	if id != -1 {
		nodelist = &global.GlobalApacheAgents[id]
	} else {
		nodelist = &global.GlobalBestAgent
	}
	ping, err := utils.PingNode(nodelist)
	if err != nil {
		return "测试失败"
	} else {
		result := fmt.Sprintf("%v", ping)
		return result
	}
}
