package main

import (
	"bufio"
	"fmt"
	"os"
	"speedtest-gd/global"
	"speedtest-gd/runtime"
	"speedtest-gd/utils"
	"strconv"
	"strings"
)

func main() {
	err := runtime.InitGlobal()
	if err != nil {
		fmt.Println("[x] 在线配置加载失败:", err)
		return
	}

	fmt.Println("IP地址:", global.GlobalClientInfo.HostIP)
	fmt.Println("城市名称:", global.GlobalClientInfo.City)
	fmt.Println("城市ID:", global.GlobalClientInfo.CityID)
	fmt.Println("运营商:", global.GlobalClientInfo.ISP)

	LocalPath := "local"
	localAgents, err := runtime.LoadAllLocalApacheAgents(LocalPath)
	if err != nil {
		fmt.Println("[x] 本地配置加载失败:", err)
	} else {
		global.GlobalApacheAgents = utils.MergeUnique(global.GlobalApacheAgents, localAgents)
	}

	fmt.Println("\n⚡️正在测试最佳节点...\n")
	bestNode, err := runtime.SelectBestNode()
	if err != nil {
		fmt.Println("[x]", err)
		return
	}
	fmt.Println("✅ 最优节点选择成功！\n")
	fmt.Printf("名称: %s\n", bestNode.Name)
	fmt.Printf("信息: %s\n", bestNode.Description)
	NewBandWidth := utils.BandwidthToGbps(bestNode.BandWidth)
	fmt.Printf("带宽限制: %d Gbps\n", int(NewBandWidth))
	choice := promptUserChoice()
	switch choice {
	case 1:
		singleDResult := runtime.SingleThreadTest(*bestNode)
		printSpeedTestResult(singleDResult)
		singleUResult := runtime.SingleThreadUploadTest(*bestNode)
		printSpeedTestResult(singleUResult)
	case 2:
		fmt.Print("请输入并发线程数（推荐 4/8/16）: ")
		threads := readIntInput()
		multiDResult := runtime.MultiThreadTest(*bestNode, threads)
		printSpeedTestResult(multiDResult)
		multiUResult := runtime.MultiThreadUploadTest(*bestNode, threads)
		printSpeedTestResult(multiUResult)
	case 3:
		fmt.Println("[!] 多节点测速不支持上传测速")
		fmt.Println("【正在进行多节点测速】")
		nodeResults := runtime.MultiNodeTest(*bestNode)
		for _, res := range nodeResults {
			printSpeedTestResult(res)
		}
		summarizeSpeed(nodeResults)
	default:
		fmt.Println("[!] 无效的选择，请重新运行程序并输入正确的数字 (1/2/3)。")
	}
}

func promptUserChoice() int {
	fmt.Println(`
请选择要执行的测速方式：
1. 单线程测速 (1 thread)
2. 多线程测速 (自定义线程数)
3. 多节点测速`)
	fmt.Print("请输入选项 (1/2/3): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > 3 {
		return -1
	}
	return choice
}

func readIntInput() int {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	i, err := strconv.Atoi(input)
	if err != nil {
		fmt.Println("[警告] 输入不是合法数字，默认使用 4 线程")
		return 4
	}
	return i
}

func printSpeedTestResult(res global.SpeedTestResult) {

	fmt.Printf(`
==============================
✅ 节点名称   : %s
📍 IP 地址    : %s
📶 测速速度   : %s
🕑 总时长     : %d ms (%.2f s)
🧵 并发线程数 : %d
==============================
`,
		res.NodeName,
		res.HostIP,
		utils.ReadableSize(res.SpeedKBps),
		res.DurationMs,
		float64(res.DurationMs)/1000.0,
		res.Threads,
	)
}

func summarizeSpeed(results []global.SpeedTestResult) {
	var totalSpeedKBps float64
	for _, res := range results {
		totalSpeedKBps += res.SpeedKBps
	}

	fmt.Printf("\n📤【多节点速度】 %s\n", utils.ReadableSize(totalSpeedKBps))
	fmt.Printf("📍 节点数量: %d\n", len(results))
}
