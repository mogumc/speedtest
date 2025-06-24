package runtime

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"speedtest-gd/global"
	"speedtest-gd/utils"
)

type HelloRequest struct {
	NeedToken                bool   `json:"needToken"`
	Manufacturer             string `json:"manufacturer"`
	NeedClient               bool   `json:"needClient"`
	NeedNatIp                bool   `json:"needNatIp"`
	NeedNetIq                bool   `json:"needNetIq"`
	NeedPretreatment         bool   `json:"needPretreatment"`
	NeedReference            bool   `json:"needReference"`
	NeedReferenceApacheAgent bool   `json:"needReferenceApacheAgent"`
	NeedWebPlugin            bool   `json:"needWebPlugin"`
}

func InitGlobal() error {
	respBytes, err := SendHelloRequest()
	if err != nil {
		return fmt.Errorf("发送 Hello 请求失败: %v", err)
	}

	var respMap map[string]interface{}
	if err := json.Unmarshal(respBytes, &respMap); err != nil {
		return fmt.Errorf("解析 JSON 失败: %v", err)
	}

	client, ok := respMap["client"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("client 字段解析失败")
	}

	if hostIp, ok := client["hostIp"].(string); ok {
		global.GlobalClientInfo.HostIP = hostIp
	}

	location, ok := client["location"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("location 字段解析失败")
	}
	if shortName, ok := location["shortName"].(string); ok {
		global.GlobalClientInfo.City = shortName
	}
	if id, ok := location["id"].(float64); ok {
		global.GlobalClientInfo.CityID = int(id)
	}

	operator, ok := client["operator"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("operator 字段解析失败")
	}
	if name, ok := operator["name"].(string); ok {
		global.GlobalClientInfo.ISP = name
	}
	if ispid, ok := operator["id"].(float64); ok {
		global.GlobalClientInfo.ISPID = int(ispid)
	}

	refAgentsInterface, ok := respMap["referenceApacheAgents"].([]interface{})
	if !ok || refAgentsInterface == nil {
		return fmt.Errorf("referenceApacheAgents 缺失或不是一个数组")
	}

	for _, item := range refAgentsInterface {
		agentMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		hostIP, _ := agentMap["hostIp"].(string)

		location := 0
		if v, ok := agentMap["location"].(float64); ok {
			location = int(v)
		}

		name, _ := agentMap["name"].(string)

		description, _ := agentMap["description"].(string)

		operator := 0
		if v, ok := agentMap["operator"].(float64); ok {
			operator = int(v)
		}

		blockSize := int64(0)
		if v, ok := agentMap["blockSize"].(float64); ok {
			blockSize = int64(v)
		}

		bandwidth := int64(0)
		if v, ok := agentMap["bandwidth"].(float64); ok {
			bandwidth = int64(v)
		}

		protocol, _ := agentMap["protocol"].(string)

		global.GlobalApacheAgents = append(global.GlobalApacheAgents, global.ApacheAgent{
			HostIP:       hostIP,
			Location:     location,
			Name:         name,
			Operator:     operator,
			BlockSize:    blockSize,
			BandWidth:    bandwidth,
			Protocol:     protocol,
			Description:  description,
			DownloadPath: "speed/100.data",
			UploadPath:   "speed/100000.data",
		})

	}

	return nil
}

func SendHelloRequest() ([]byte, error) {
	url := "https://speed.gd.cn/hello"

	requestBody := HelloRequest{
		NeedToken:                false,
		Manufacturer:             "vixtel",
		NeedClient:               true,
		NeedNatIp:                true,
		NeedNetIq:                true,
		NeedPretreatment:         true,
		NeedReference:            true,
		NeedReferenceApacheAgent: true,
		NeedWebPlugin:            true,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko)",
	}

	response, err := utils.HTTPPost(url, headers, bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("HTTP POST request failed: %v", err)
	}

	return response, nil
}

func LoadLocalApacheAgentsFromFile(filePath string) ([]global.ApacheAgent, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法读取文件：%v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("JSON 解析失败：%v", err)
	}
	arr, ok := result["referenceApacheAgents"].([]interface{})
	if !ok || arr == nil {
		return nil, fmt.Errorf("referenceApacheAgents 缺失或不是一个数组")
	}
	var agents []global.ApacheAgent

	for _, item := range arr {
		agentMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		hostIP, _ := agentMap["hostIp"].(string)

		location := 0
		if v, ok := agentMap["location"].(float64); ok {
			location = int(v)
		}

		name, _ := agentMap["name"].(string)

		description, _ := agentMap["description"].(string)

		operator := 0
		if v, ok := agentMap["operator"].(float64); ok {
			operator = int(v)
		}

		blockSize := int64(0)
		if v, ok := agentMap["blockSize"].(float64); ok {
			blockSize = int64(v)
		}

		bandwidth := int64(0)
		if v, ok := agentMap["bandwidth"].(float64); ok {
			bandwidth = int64(v)
		}

		protocol, _ := agentMap["protocol"].(string)

		downloadpath, _ := agentMap["downloadpath"].(string)
		uploadpath, _ := agentMap["uploadpath"].(string)

		agents = append(agents, global.ApacheAgent{
			HostIP:       hostIP,
			Location:     location,
			Name:         name,
			Operator:     operator,
			BlockSize:    blockSize,
			BandWidth:    bandwidth,
			Protocol:     protocol,
			Description:  description,
			DownloadPath: downloadpath,
			UploadPath:   uploadpath,
		})

	}
	return agents, nil
}

func LoadAllLocalApacheAgents(dirPath string) ([]global.ApacheAgent, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("无法读取目录：%v", err)
	}
	var allAgents []global.ApacheAgent
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}
		filePath := filepath.Join(dirPath, file.Name())
		fmt.Printf("[!] 正在解析 %s\n", filePath)
		agents, err := LoadLocalApacheAgentsFromFile(filePath)
		if err != nil {
			fmt.Printf("[x] 文件 %q 无效（%v），已跳过。\n", filePath, err)
			continue
		}
		allAgents = append(allAgents, agents...)
	}
	return allAgents, nil
}
