package runtimes

import (
	"encoding/json"
	"fmt"
	"speedtest-gd/global"
	"speedtest-gd/utils"
	"time"
)

type PingResult struct {
	Agent *global.ApacheAgent
	Delay time.Duration
	Err   error
}

func SelectBestNode() (*global.ApacheAgent, error) {
	cityID := global.GlobalClientInfo.CityID
	NetID := global.GlobalClientInfo.ISPID
	var matching []*global.ApacheAgent
	for idx := range global.GlobalApacheAgents {
		agent := &global.GlobalApacheAgents[idx]
		if agent.Location == cityID && agent.Operator == NetID {
			matching = append(matching, agent)
		}
	}

	if len(matching) > 0 {
		return matching[0], nil
	}

	for idx := range global.GlobalApacheAgents {
		agent := &global.GlobalApacheAgents[idx]
		if agent.Location == cityID {
			matching = append(matching, agent)
		}
	}

	if len(matching) > 0 {
		return matching[0], nil
	}

	resultsChan := make(chan PingResult, len(global.GlobalApacheAgents))
	for idx := range global.GlobalApacheAgents {
		go func(agent *global.ApacheAgent) {
			delay, err := utils.PingNode(agent)
			if err != nil {
				fmt.Printf("[x] 测试节点 %s 失败: %v\n", agent.Name, err)
			}
			resultsChan <- PingResult{
				Agent: agent,
				Delay: delay,
				Err:   err,
			}
		}(&global.GlobalApacheAgents[idx])
	}
	var best *global.ApacheAgent
	var minDelay time.Duration = time.Minute
	for i := 0; i < cap(resultsChan); i++ {
		result := <-resultsChan
		if result.Err != nil {
			continue
		}
		if result.Delay < minDelay {
			minDelay = result.Delay
			best = result.Agent
		}
	}
	if best != nil {
		return best, nil
	}
	return nil, fmt.Errorf("无法获取可用节点")
}

func ShowAllNode(Agents []global.ApacheAgent) string {
	jsonData, err := json.Marshal(Agents)
	if err != nil {
		Error := fmt.Sprintf("%v", err)
		return Error
	} else {
		return string(jsonData)
	}
}
