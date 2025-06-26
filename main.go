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
		fmt.Println("[x] 在线配置加载失败: %v", err)
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

	fmt.Printf("\n⚡️正在测试最佳节点...\n")
	bestNode, err := runtime.SelectBestNode()
	if err != nil {
		fmt.Printf("[x] %v", err)
		return
	}
	NewBandWidth := utils.BandwidthToGbps(bestNode.BandWidth)
	fmt.Printf(`
==============================
✅ 最优节点选择成功！
✅ 节点名称   : %s
🎭 描述       : %s
📍 IP 地址    : %s
🚀 最大速度   : %f Gbps
==============================
`, bestNode.Name, bestNode.Description, bestNode.HostIP, NewBandWidth)
	choice := promptUserChoice()
	switch choice {
	case 1:
		ping, err := utils.PingNode(bestNode)
		if err != nil {
			fmt.Printf("[x] 节点测试失败！请手动选择其他节点！")
			return
		}
		fmt.Printf(`
==============================
✅ 节点名称   : %s
🎭 描述       : %s
📍 IP 地址    : %s
🚀 最大速度   : %f Gbps
⚡️ 延迟       : %v
==============================
`, bestNode.Name, bestNode.Description, bestNode.HostIP, NewBandWidth, ping)
		singleDResult := runtime.SingleThreadTest(*bestNode)
		printSpeedTestResult(singleDResult)
		singleUResult := runtime.SingleThreadUploadTest(*bestNode)
		printSpeedTestResult(singleUResult)
	case 2:
		ping, err := utils.PingNode(bestNode)
		if err != nil {
			fmt.Printf("[x] 节点测试失败！请手动选择其他节点！")
			return
		}
		fmt.Printf(`
==============================
✅ 节点名称   : %s
🎭 描述       : %s
📍 IP 地址    : %s
🚀 最大速度   : %f Gbps
⚡️ 延迟       : %v
==============================
`, bestNode.Name, bestNode.Description, bestNode.HostIP, NewBandWidth, ping)
		fmt.Print("请输入并发线程数（推荐 4/8/16）: ")
		threads := readIntInput()
		multiDResult := runtime.MultiThreadTest(*bestNode, threads)
		printSpeedTestResult(multiDResult)
		multiUResult := runtime.MultiThreadUploadTest(*bestNode, threads)
		printSpeedTestResult(multiUResult)
	case 3:
		fmt.Println("[!] 多节点测速暂不支持上传测速")
		fmt.Println("【正在进行多节点测速】")
		nodeResults := runtime.MultiNodeTest(*bestNode)
		for _, res := range nodeResults {
			printSpeedTestResult(res)
		}
		summarizeSpeed(nodeResults)
	case 4:
		fmt.Print("\n")
		runtime.ShowAllNode(global.GlobalApacheAgents)
		fmt.Print("\n 请输入选择的节点编号: ")
		idx := readIntInput()
		ping, err := utils.PingNode(&global.GlobalApacheAgents[idx])
		if err != nil && ping == 0 {
			fmt.Printf("[x] 节点测试失败！请手动选择其他节点！")

		}
		NewChoseBandWidth := utils.BandwidthToGbps(global.GlobalApacheAgents[idx].BandWidth)
		fmt.Printf(`
==============================
✅ 节点名称   : %s
🎭 描述       : %s
📍 IP 地址    : %s
🚀 最大速度   : %f Gbps
⚡️ 延迟       : %v
==============================
`, global.GlobalApacheAgents[idx].Name, global.GlobalApacheAgents[idx].Description, global.GlobalApacheAgents[idx].HostIP, NewChoseBandWidth, ping)
		fmt.Print("请输入并发线程数（推荐 4/8/16）: ")
		threads := readIntInput()
		multiDResult := runtime.MultiThreadTest(global.GlobalApacheAgents[idx], threads)
		printSpeedTestResult(multiDResult)
		multiUResult := runtime.MultiThreadUploadTest(global.GlobalApacheAgents[idx], threads)
		printSpeedTestResult(multiUResult)
	default:
		fmt.Println("[!] 无效的选择，请重新运行程序并输入正确的数字 (1/2/3/4)。")
	}
}

func promptUserChoice() int {
	fmt.Println(`
请选择要执行的测速方式：
1. 单线程测速 (1 thread)
2. 多线程测速 (自定义线程数)
3. 多节点测速
4. 自选节点测速`)
	fmt.Print("请输入选项 (1/2/3/4): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > 4 {
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
		fmt.Println("[!] 输入不是合法数字，默认使用 4 线程")
		return 4
	}
	if i > 64 {
		fmt.Println("[!] 最大线程 64 线程,将默认使用最大线程")
		return 64
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
