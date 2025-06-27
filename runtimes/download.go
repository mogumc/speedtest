package runtimes

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"speedtest-gd/global"
	"speedtest-gd/utils"
	"strconv"
	"time"
)

func DownloadTestWithURL(url string, ThreadID int) (speedKBps float64, durationMs int64, totaldata float64) {
	client := &http.Client{
		Timeout: global.MaxTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("[x] 创建请求失败: %v\n", err)
		return
	}
	req.Header.Set("User-Agent", global.UserAgent)
	var totalBytes int64
	var reqCount int
	var firstStart, lastEnd time.Time
	stopTimer := time.NewTimer(global.TestDuration)
	defer stopTimer.Stop()
	for {
		select {
		case <-stopTimer.C:
			goto ExitLoop
		default:
		}
		start := time.Now()
		if firstStart.IsZero() {
			firstStart = start
		}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("[x] 请求失败 %s | 错误：%v\n", url, err)
			continue
		}
		expectedSize, isChunked := parseContentHeader(resp)
		bytesRead, err := io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("[x] 读取响应体失败 %s | 错误：%v\n", url, err)
			continue
		}
		reqCount++
		elapsed := time.Since(start).Seconds()
		var currentBytes int64
		if !isChunked && expectedSize > 0 {
			currentBytes = expectedSize
		} else {
			currentBytes = bytesRead
		}
		totalBytes += currentBytes
		lastEnd = time.Now()
		speedKBps = float64(currentBytes) / 1024 / elapsed
		global.GlobalSpeed.Mutex.Lock()
		threadID := string(ThreadID)
		global.GlobalSpeed.ThreadDSpeeds[threadID] = speedKBps
		global.GlobalSpeed.TotalDData += float64(currentBytes)
		global.GlobalSpeed.RequestCount++
		global.GlobalSpeed.LastUpdate = lastEnd
		global.GlobalSpeed.Mutex.Unlock()
	}
ExitLoop:
	if reqCount < 1 {
		fmt.Printf("[x] 测试失败!当前节点不可用或速度过慢!\n")
		return
	}
	finalDuration := lastEnd.Sub(firstStart)
	speedKBps = float64(totalBytes) / 1024 / finalDuration.Seconds()
	durationMs = finalDuration.Milliseconds()
	dataMB := float64(totalBytes) / (1024 * 1024)
	fmt.Printf(
		"[✔] 测试结束 URL=%s | %.2f KB/s | 总 %d 秒 | 总数据 %.2f MB | 成功下载次数 %d 次请求\n",
		url, speedKBps, int(finalDuration.Round(time.Second).Seconds()), dataMB, reqCount,
	)
	return speedKBps, durationMs, dataMB
}

func parseContentHeader(resp *http.Response) (expectedSize int64, isChunked bool) {
	if cl := resp.Header.Get("Content-Length"); cl != "" {
		var err error
		expectedSize, err = strconv.ParseInt(cl, 10, 64)
		if err != nil || expectedSize < 0 {
			return 0, true
		}
		return expectedSize, false
	}
	return 0, true
}
func SingleThreadTest(bestNode global.ApacheAgent) global.SpeedTestResult {
	url := fmt.Sprintf("%s://%s/%s", bestNode.Protocol, bestNode.HostIP, bestNode.DownloadPath)
	fmt.Printf("[+] 单线程下载测速开始 URL=%s\n", url)
	speed, duration, totaldata := DownloadTestWithURL(url, 1)
	return global.SpeedTestResult{
		NodeName:   bestNode.Name,
		HostIP:     bestNode.HostIP,
		SpeedKBps:  speed,
		DurationMs: duration,
		Threads:    1,
		TotalData:  totaldata,
	}
}

func MultiThreadTest(bestNode global.ApacheAgent, threadCount int) global.SpeedTestResult {
	url := fmt.Sprintf("%s://%s/%s", bestNode.Protocol, bestNode.HostIP, bestNode.DownloadPath)
	fmt.Printf("[+] 多线程下载测速开始 URL=%s Thread=%d\n", url, threadCount)
	type result struct {
		speed     float64
		duration  int64
		totaldata float64
	}
	resultsChan := make(chan result, threadCount)

	for i := 0; i < threadCount; i++ {
		go func() {
			speed, duration, totaldata := DownloadTestWithURL(url, i)
			resultsChan <- result{speed: speed, duration: duration, totaldata: totaldata}
		}()
	}

	var totalSpeed float64
	var smdata float64
	var durations []int64
	count := 0
	for r := range resultsChan {
		totalSpeed += r.speed
		smdata += r.totaldata
		durations = append(durations, r.duration)
		count++
		if count >= threadCount {
			break
		}
	}

	var avgDuration int64
	if count > 0 {
		var sumDur int64
		for _, dur := range durations {
			sumDur += dur
		}
		avgDuration = sumDur / int64(count)
	}

	fmt.Printf("[+] 多线程下载总结 速度 %s 数据 %.2f MB\n",
		utils.ReadableSize(totalSpeed), smdata)

	return global.SpeedTestResult{
		NodeName:   bestNode.Name,
		HostIP:     bestNode.HostIP,
		SpeedKBps:  totalSpeed,
		DurationMs: avgDuration,
		Threads:    threadCount,
		TotalData:  smdata,
	}
}
