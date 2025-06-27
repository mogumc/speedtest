package service

import (
	"fmt"
	"speedtest-gd/global"
	"speedtest-gd/runtimes"
	"speedtest-gd/utils"
	"sync"
	"time"
)

func startTest(id []int, threads int, mode int) {
	CleanupSpeed()
	StartGlobalSpeedUpdater()
	if len(id) == 0 {
		muiltnodetest(global.GlobalBestAgent, threads, mode)
		global.GlobalSpeed.Mutex.Lock()
		global.GlobalSpeed.Is_done = 1
		global.GlobalSpeed.Mutex.Unlock()
	} else if len(id) == 1 {
		muiltnodetest(global.GlobalApacheAgents[id[0]], threads, mode)
		global.GlobalSpeed.Mutex.Lock()
		global.GlobalSpeed.Is_done = 1
		global.GlobalSpeed.Mutex.Unlock()
	} else if len(id) > 1 {
		var downresults []global.SpeedTestResult
		var upresults []global.SpeedTestResult
		if mode == 1 || mode == 0 {
			global.GlobalSpeed.RequestCount = 0
			resultChan := make(chan global.SpeedTestResult, len(id))
			var wg sync.WaitGroup
			wg.Add(len(id))
			for _, idx := range id {
				go func(agent global.ApacheAgent) {
					defer wg.Done()
					res := runtimes.SingleThreadTest(agent)
					resultChan <- res
				}(global.GlobalApacheAgents[idx])
			}
			wg.Wait()
			close(resultChan)
			for res := range resultChan {
				downresults = append(downresults, res)
			}
			SpeedKBps, TotalData := summarizeSpeed(downresults)
			global.GlobalSpeed.Mutex.Lock()
			global.GlobalSpeed.DownSpeedKBps = SpeedKBps
			global.GlobalSpeed.TotalDData = TotalData * 1024 * 1024
			global.GlobalSpeed.Mutex.Unlock()
		}
		if mode == 2 || mode == 0 {
			global.GlobalSpeed.RequestCount = 0
			resultChan := make(chan global.SpeedTestResult, len(id))
			var wg sync.WaitGroup
			wg.Add(len(id))
			for _, idx := range id {
				go func(agent global.ApacheAgent) {
					defer wg.Done()
					res := runtimes.SingleThreadUploadTest(agent)
					resultChan <- res
				}(global.GlobalApacheAgents[idx])
			}
			wg.Wait()
			close(resultChan)
			for res := range resultChan {
				upresults = append(upresults, res)
			}
			global.GlobalSpeed.Mutex.Lock()
			SpeedKBps, TotalData := summarizeSpeed(upresults)
			global.GlobalSpeed.UpSpeedKBps = SpeedKBps
			global.GlobalSpeed.TotalUData = TotalData * 1024 * 1024
			global.GlobalSpeed.Mutex.Unlock()
		}
		global.GlobalSpeed.Mutex.Lock()
		global.GlobalSpeed.Is_done = 1
		global.GlobalSpeed.Mutex.Unlock()
	} else {
		global.GlobalSpeed.Mutex.Lock()
		global.GlobalSpeed.Is_done = 1
		global.GlobalSpeed.Mutex.Unlock()
		return
	}
}

func muiltnodetest(agent global.ApacheAgent, threads, mode int) error {
	NewBandWidth := utils.BandwidthToGbps(agent.BandWidth)
	ping, err := utils.PingNode(&agent)
	if err != nil {
		return err
	}
	fmt.Printf(`
==============================
✅ 节点名称   : %s
🎭 描述       : %s
📍 IP 地址    : %s
🚀 最大速度   : %f Gbps
⚡️ 延迟       : %v
==============================
`, agent.Name, agent.Description, agent.HostIP, NewBandWidth, ping)
	if mode == 1 || mode == 0 {
		multiDResult := runtimes.MultiThreadTest(agent, threads)
		global.GlobalSpeed.Mutex.Lock()
		global.GlobalSpeed.DownSpeedKBps = multiDResult.SpeedKBps
		global.GlobalSpeed.TotalDData = multiDResult.TotalData * 1024 * 1024
		global.GlobalSpeed.Mutex.Unlock()
	}
	if mode == 2 || mode == 0 {
		multiUResult := runtimes.MultiThreadUploadTest(agent, threads)
		global.GlobalSpeed.Mutex.Lock()
		global.GlobalSpeed.UpSpeedKBps = multiUResult.SpeedKBps
		global.GlobalSpeed.TotalUData = multiUResult.TotalData * 1024 * 1024
		global.GlobalSpeed.Mutex.Unlock()
	}
	return nil
}

func summarizeSpeed(results []global.SpeedTestResult) (float64, float64) {
	var totalSpeedKBps float64
	var totalData float64
	for _, res := range results {
		totalSpeedKBps += res.SpeedKBps
		totalData += res.TotalData
	}
	return totalSpeedKBps, totalData
}

func StartGlobalSpeedUpdater() {
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for {
			global.GlobalSpeed.Mutex.Lock()
			isDone := global.GlobalSpeed.Is_done
			global.GlobalSpeed.Mutex.Unlock()
			if isDone == 1 {
				ticker.Stop()
				return
			}
			<-ticker.C
			global.GlobalSpeed.Mutex.Lock()
			var downTotal, upTotal float64
			for _, speed := range global.GlobalSpeed.ThreadDSpeeds {
				downTotal += speed
			}
			for _, speed := range global.GlobalSpeed.ThreadUSpeeds {
				upTotal += speed
			}
			global.GlobalSpeed.DownSpeedKBps = downTotal
			global.GlobalSpeed.UpSpeedKBps = upTotal
			global.GlobalSpeed.LastUpdate = time.Now()
			global.GlobalSpeed.Mutex.Unlock()
		}
	}()
}

func CleanupSpeed() {
	global.GlobalSpeed.Mutex.Lock()
	defer global.GlobalSpeed.Mutex.Unlock()

	global.GlobalSpeed.DownSpeedKBps = 0
	global.GlobalSpeed.UpSpeedKBps = 0
	global.GlobalSpeed.TotalDData = 0
	global.GlobalSpeed.TotalUData = 0
	global.GlobalSpeed.RequestCount = 0
	global.GlobalSpeed.LastUpdate = time.Now()
	global.GlobalSpeed.Is_done = 0

	for k := range global.GlobalSpeed.ThreadDSpeeds {
		delete(global.GlobalSpeed.ThreadDSpeeds, k)
	}
	for k := range global.GlobalSpeed.ThreadUSpeeds {
		delete(global.GlobalSpeed.ThreadUSpeeds, k)
	}
}
