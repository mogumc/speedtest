package runtime

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"speedtest-gd/global"
	"speedtest-gd/utils"
	"strconv"
	"sync"
	"time"
)

func DownloadTestWithURL(url string, _ int) (speedKBps float64, durationMs int64, totaldata float64) {
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
		// 关键：只统计bytes read而不保存文件
		bytesRead, err := io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("[x] 读取响应体失败 %s | 错误：%v\n", url, err)
			continue
		}
		reqCount++
		if !isChunked && expectedSize > 0 {
			totalBytes += expectedSize
		} else {
			totalBytes += bytesRead
		}
		lastEnd = time.Now()
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
			speed, duration, totaldata := DownloadTestWithURL(url, threadCount)
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

func SwitchNodesForMultiNodeTesting(excludeNode global.ApacheAgent) []global.ApacheAgent {
	var candidates []global.ApacheAgent

	for _, node := range global.GlobalApacheAgents {
		candidates = append(candidates, node)
	}

	if len(candidates) > 2 {
		candidates = candidates[:2]
	}

	candidates = append(candidates, excludeNode)

	return candidates
}

func MultiNodeTest(baseNode global.ApacheAgent) []global.SpeedTestResult {
	extraNodes := SwitchNodesForMultiNodeTesting(baseNode)
	if len(extraNodes) == 0 {
		return []global.SpeedTestResult{}
	}
	resultChan := make(chan global.SpeedTestResult, len(extraNodes))
	var results []global.SpeedTestResult
	var wg sync.WaitGroup
	wg.Add(len(extraNodes))
	for _, node := range extraNodes {
		go func(agent global.ApacheAgent) {
			defer wg.Done()
			res := SingleThreadTest(agent)
			resultChan <- res
		}(node)
	}
	wg.Wait()
	close(resultChan)
	for res := range resultChan {
		results = append(results, res)
	}
	return results
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
