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

	} else if len(id) == 1 {
		muiltnodetest(global.GlobalApacheAgents[id[0]], threads, mode)
	} else if len(id) > 1 {
		var downresults []global.SpeedTestResult
		var upresults []global.SpeedTestResult
		var DownSpeedKBps float64 = 0
		var TotalDData float64 = 0
		var UpSpeedKBps float64 = 0
		var TotalUData float64 = 0
		if mode == 1 || mode == 0 {
			global.GlobalSpeed.Mutex.Lock()
			global.GlobalSpeed.RequestCount = 0
			global.GlobalSpeed.Mutex.Unlock()
			resultChan := make(chan global.SpeedTestResult, len(id))
			var wg sync.WaitGroup
			wg.Add(len(id))
			for _, idx := range id {
				go func(agent global.ApacheAgent) {
					defer wg.Done()
					res := runtimes.MultiThreadTest(agent, 1)
					resultChan <- res
				}(global.GlobalApacheAgents[idx])
			}
			wg.Wait()
			close(resultChan)
			for res := range resultChan {
				downresults = append(downresults, res)
			}
			SpeedKBps, TotalData := summarizeSpeed(downresults)
			DownSpeedKBps = SpeedKBps
			TotalDData = TotalData * 1024 * 1024
		}
		if mode == 2 || mode == 0 {
			global.GlobalSpeed.Mutex.Lock()
			global.GlobalSpeed.RequestCount = 0
			global.GlobalSpeed.Mutex.Unlock()
			resultChan := make(chan global.SpeedTestResult, len(id))
			var wg sync.WaitGroup
			wg.Add(len(id))
			for _, idx := range id {
				go func(agent global.ApacheAgent) {
					defer wg.Done()
					res := runtimes.MultiThreadUploadTest(agent, 1)
					resultChan <- res
				}(global.GlobalApacheAgents[idx])
			}
			wg.Wait()
			close(resultChan)
			for res := range resultChan {
				upresults = append(upresults, res)
			}
			SpeedKBps, TotalData := summarizeSpeed(upresults)
			UpSpeedKBps = SpeedKBps
			TotalUData = TotalData * 1024 * 1024
		}
		global.GlobalSpeed.Mutex.Lock()
		global.GlobalSpeed.Is_done = 1
		global.GlobalSpeed.DownSpeedKBps = DownSpeedKBps
		global.GlobalSpeed.TotalDData = TotalDData
		global.GlobalSpeed.UpSpeedKBps = UpSpeedKBps
		global.GlobalSpeed.TotalUData = TotalUData
		global.GlobalSpeed.Mutex.Unlock()
	} else {
		global.GlobalSpeed.Mutex.Lock()
		global.GlobalSpeed.Is_done = 1
		global.GlobalSpeed.DownSpeedKBps = 0
		global.GlobalSpeed.TotalDData = 0
		global.GlobalSpeed.UpSpeedKBps = 0
		global.GlobalSpeed.TotalUData = 0
		global.GlobalSpeed.Mutex.Unlock()
		return
	}
}

func muiltnodetest(agent global.ApacheAgent, threads, mode int) error {
	var DownSpeedKBps float64 = 0
	var TotalDData float64 = 0
	var UpSpeedKBps float64 = 0
	var TotalUData float64 = 0
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
		DownSpeedKBps = multiDResult.SpeedKBps
		TotalDData = multiDResult.TotalData * 1024 * 1024
	}
	if mode == 2 || mode == 0 {
		global.GlobalSpeed.UpStartAt = time.Now()
		multiUResult := runtimes.MultiThreadUploadTest(agent, threads)
		UpSpeedKBps = multiUResult.SpeedKBps
		TotalUData = multiUResult.TotalData * 1024 * 1024
	}
	global.GlobalSpeed.Mutex.Lock()
	global.GlobalSpeed.Is_done = 1
	global.GlobalSpeed.DownSpeedKBps = DownSpeedKBps
	global.GlobalSpeed.TotalDData = TotalDData
	global.GlobalSpeed.UpSpeedKBps = UpSpeedKBps
	global.GlobalSpeed.TotalUData = TotalUData
	global.GlobalSpeed.Mutex.Unlock()
	return nil
}

func summarizeSpeed(results []global.SpeedTestResult) (float64, float64) {
	var totalSpeedKBps float64
	var totalData float64
	for _, res := range results {
		totalSpeedKBps += res.SpeedKBps
		totalData += res.TotalData
	}

	fmt.Printf("\n📤【多节点速度】 %s\n", utils.ReadableSize(totalSpeedKBps))
	fmt.Printf("📍 成功节点数量: %d\n", len(results))
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
			if global.GlobalSpeed.DownDone != 1 {
				downTotal = global.GlobalSpeed.TotalDData
				downStartAt := global.GlobalSpeed.DownStartAt
				DownSpeedKBps := calculateSpeedInKBps(downTotal, downStartAt)
				global.GlobalSpeed.DownSpeedKBps = DownSpeedKBps
			}
			if global.GlobalSpeed.UpDone != 1 {
				upTotal = global.GlobalSpeed.TotalUData
				upStartAt := global.GlobalSpeed.UpStartAt
				UpSpeedKBps := calculateSpeedInKBps(upTotal, upStartAt)
				global.GlobalSpeed.UpSpeedKBps = UpSpeedKBps
			}
			global.GlobalSpeed.LastUpdate = time.Now()
			global.GlobalSpeed.Mutex.Unlock()
		}
	}()
}

func calculateSpeedInKBps(current float64, startTime time.Time) float64 {
	elapsed := time.Since(startTime).Seconds()
	if elapsed <= 0 {
		return 0
	}
	speedInBytes := current / elapsed
	speedInKB := speedInBytes / 1024.0
	return speedInKB
}

func CleanupSpeed() {
	global.GlobalSpeed.Mutex.Lock()
	defer global.GlobalSpeed.Mutex.Unlock()

	global.GlobalSpeed.DownSpeedKBps = 0
	global.GlobalSpeed.UpSpeedKBps = 0
	global.GlobalSpeed.TotalDData = 0
	global.GlobalSpeed.TotalUData = 0
	global.GlobalSpeed.RequestCount = 0
	global.GlobalSpeed.DownStartAt = time.Now()
	global.GlobalSpeed.UpStartAt = time.Now()
	global.GlobalSpeed.LastUpdate = time.Now()
	global.GlobalSpeed.Is_done = 0
}
