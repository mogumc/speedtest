// runtime/upload.go

package runtime

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"speedtest-gd/global"
	"speedtest-gd/utils"
)

func GenerateRandomPayload() []byte {
	payload := make([]byte, global.UploadBlockSize)
	rand.Read(payload)
	return payload
}

type TrackedReader struct {
	r     io.Reader
	Total int64
	mu    sync.Mutex
}

func (t *TrackedReader) Read(p []byte) (int, error) {
	n, err := t.r.Read(p)
	t.mu.Lock()
	t.Total += int64(n)
	t.mu.Unlock()
	return n, err
}

func UploadTestWithURL(url string, duration time.Duration, threads int) (result global.SpeedTestResult) {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 10,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	body := GenerateRandomPayload()
	var totalBytesSent int64 = 0
	var count int = 0
	var totalTime time.Duration = 0
	startTime := time.Now()
	reqCount := 0
	for time.Since(startTime) < duration {
		reqCount++
		reqStart := time.Now()
		payloadReader := bytes.NewReader(body)
		trackedReader := &TrackedReader{
			r: payloadReader,
		}
		req, _ := http.NewRequest("POST", url, trackedReader)
		req.Header.Set("User-Agent", global.UserAgent)
		req.Header.Set("Content-Type", "application/octet-stream")
		req.ContentLength = int64(payloadReader.Size())
		ctx, cancel := context.WithTimeout(context.Background(), global.MaxTimeout)
		defer cancel()
		req = req.WithContext(ctx)
		uploadStart := time.Now()
		resp, err := client.Do(req)
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Printf("[!] 上传超时放弃 URL=%s | 用时 %.2f 秒\n", url, time.Since(reqStart).Seconds())
			continue
		}
		if err != nil {
			fmt.Printf("[x] 请求失败 %v\n", err)
			continue
		}
		_ = resp.Body.Close()
		actualSent := trackedReader.Total
		uploadDuration := time.Since(uploadStart)
		totalTime += uploadDuration
		totalBytesSent += actualSent
		count++
	}
	if count < 1 {
		fmt.Printf("[x] 测试失败!当前节点不可用!\n")
		return
	}
	elapsed := time.Since(startTime)
	totalSeconds := elapsed.Seconds()
	totalKBSent := float64(totalBytesSent) / 1024
	speedKBps := totalKBSent / totalSeconds
	fmt.Printf("[✔] 测试结束 URL=%s | 总 %.2f KB/s | 总 %.2f 秒 | 总数据 %.2f MB | 成功上传次数 %d\n",
		url, speedKBps, totalSeconds, totalKBSent/(1024), count)
	result = global.SpeedTestResult{
		NodeName:   extractNodeName(url),
		HostIP:     extractHostIP(url),
		SpeedKBps:  speedKBps,
		DurationMs: elapsed.Milliseconds(),
		Threads:    threads,
		TotalData:  totalKBSent / (1024),
	}
	return result
}

func SingleThreadUploadTest(bestNode global.ApacheAgent) global.SpeedTestResult {
	url := fmt.Sprintf("%s://%s/%s", bestNode.Protocol, bestNode.HostIP, bestNode.UploadPath)
	fmt.Printf("[+] 单线程上传测速开始 URL=%s\n", url)
	return UploadTestWithURL(url, global.TestDuration, 1)
}

func MultiThreadUploadTest(bestNode global.ApacheAgent, threadCount int) global.SpeedTestResult {
	url := fmt.Sprintf("%s://%s/%s", bestNode.Protocol, bestNode.HostIP, bestNode.UploadPath)
	fmt.Printf("[+] 多线程上传测速开始 URL=%s Thread=%d\n", url, threadCount)
	type resultStruct struct {
		SpeedKBps  float64
		DurationMs int64
		TotalData  float64
	}

	resultsChan := make(chan resultStruct, threadCount)

	for i := 0; i < threadCount; i++ {
		go func() {
			res := UploadTestWithURL(url, global.TestDuration, threadCount)
			resultsChan <- resultStruct{
				SpeedKBps:  res.SpeedKBps,
				DurationMs: res.DurationMs,
				TotalData:  res.TotalData,
			}
		}()
	}

	var totalSpeed, totalData float64
	var durations []int64
	count := 0
	for r := range resultsChan {
		totalSpeed += r.SpeedKBps
		durations = append(durations, r.DurationMs)
		totalData += r.TotalData
		count++
		if count >= threadCount {
			break
		}
	}

	var avgDuration int64
	if count > 0 {
		sumDur := int64(0)
		for _, dur := range durations {
			sumDur += dur
		}
		avgDuration = sumDur / int64(count)
	}

	fmt.Printf("[+] 多线程上传总结 速度 %s 数据 %.2f MB\n",
		utils.ReadableSize(totalSpeed), totalData)

	return global.SpeedTestResult{
		NodeName:   bestNode.Name,
		HostIP:     bestNode.HostIP,
		SpeedKBps:  totalSpeed,
		DurationMs: avgDuration,
		Threads:    threadCount,
		TotalData:  totalData,
	}
}

func extractNodeName(url string) string {
	for _, node := range global.GlobalApacheAgents {
		if strings.Contains(url, node.HostIP) {
			return node.Name
		}
	}
	return "Unknown Node"
}

func extractHostIP(url string) string {
	for _, node := range global.GlobalApacheAgents {
		if strings.Contains(url, node.HostIP) {
			return node.HostIP
		}
	}
	return "Unknown IP"
}
