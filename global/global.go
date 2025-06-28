package global

import (
	"sync"
	"time"
)

type ClientInfo struct {
	HostIP   string
	City     string
	District string
	ISP      string
	Country  string
	Province string
}

var GlobalClientInfo = &ClientInfo{}

type ApacheAgent struct {
	HostIP       string
	Location     int
	Name         string
	Operator     int
	BlockSize    int64
	BandWidth    int64
	Protocol     string
	Description  string
	DownloadPath string
	UploadPath   string
}

var GlobalApacheAgents []ApacheAgent
var GlobalBestAgent ApacheAgent
var GlobalSpeed = NewSpeedTestSpeed()

var (
	UploadBlockSize = 5 * 1024 * 1024 // 5 MB
	UserAgent       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36 Edg/137.0.0.0"
	TestDuration    = 10 * time.Second
	MaxTimeout      = 30 * time.Second
)

type SpeedTestResult struct {
	NodeName   string
	HostIP     string
	SpeedKBps  float64
	DurationMs int64
	Threads    int
	TotalData  float64
}

type SpeedTestSpeed struct {
	Mutex         sync.Mutex
	DownSpeedKBps float64
	UpSpeedKBps   float64
	Threads       int
	TotalDData    float64
	TotalUData    float64
	ThreadDSpeeds map[string]float64
	ThreadUSpeeds map[string]float64
	RequestCount  int16
	LastUpdate    time.Time
	Is_done       int
}

func NewSpeedTestSpeed() *SpeedTestSpeed {
	return &SpeedTestSpeed{
		ThreadDSpeeds: make(map[string]float64),
		ThreadUSpeeds: make(map[string]float64),
	}
}
