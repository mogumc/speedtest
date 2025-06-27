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
		info += fmt.Sprintf("[x] 在线配置加载失败: %v", err)
	}
	LocalPath := "local"
	localAgents, err := runtimes.LoadAllLocalApacheAgents(LocalPath)
	if err != nil {
		info += fmt.Sprintf("[x] 本地配置加载失败: %v", err)
	} else {
		global.GlobalApacheAgents = utils.MergeUnique(global.GlobalApacheAgents, localAgents)
	}
	if len(global.GlobalApacheAgents) < 1 {
		return info
	} else {
		return "OK"
	}
}

func getInfo() string {
	datas := map[string]interface{}{
		"HostIP": global.GlobalClientInfo.HostIP,
		"City":   global.GlobalClientInfo.City,
		"CityID": global.GlobalClientInfo.CityID,
		"ISP":    global.GlobalClientInfo.ISP,
		"ISPID":  global.GlobalClientInfo.ISPID,
	}
	jsonBytes, err := json.Marshal(datas)
	if err != nil {
		return "1"
	}

	return string(jsonBytes)
}
