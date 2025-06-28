package runtimes

import (
	"encoding/base64"
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
		return fmt.Errorf("获取用户失败: %v", err)
	}
	var respMap map[string]interface{}
	if err := json.Unmarshal(respBytes, &respMap); err != nil {
		return fmt.Errorf("解析 JSON 失败: %v", err)
	}

	if hostIp, ok := respMap["ip"].(string); ok {
		global.GlobalClientInfo.HostIP = hostIp
	}

	if city, ok := respMap["city"].(string); ok {
		global.GlobalClientInfo.City = city
	}
	if district, ok := respMap["district"].(string); ok {
		global.GlobalClientInfo.District = district
	}

	if ispname, ok := respMap["isp"].(string); ok {
		global.GlobalClientInfo.ISP = ispname
	}
	if country, ok := respMap["country"].(string); ok {
		global.GlobalClientInfo.Country = country
	}
	if province, ok := respMap["province"].(string); ok {
		global.GlobalClientInfo.Province = province
	}
	return nil
}

func SendHelloRequest() ([]byte, error) {
	url := "https://api-v3.speedtest.cn/ip"

	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko)",
		"CLIENTECTYPE": "65",
	}

	response, err := utils.HTTPGet(url, headers)
	if err != nil {
		return nil, fmt.Errorf("HTTP Get request failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response JSON: %v", err)
	}

	dataField, ok := result["data"].(string)
	if !ok || dataField == "" {
		return nil, fmt.Errorf("data field not found or invalid in JSON")
	}

	key := []byte("5ECC5D62140EC099")
	iv := []byte("E63EA892A702EEAA")
	encryptedData, err := base64.StdEncoding.DecodeString(dataField)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data: %v", err)
	}
	decrypt, err := utils.AesDecrypt(encryptedData, key, iv)
	if err != nil {
		return nil, fmt.Errorf("AES decryption failed: %v", err)
	}

	return decrypt, nil

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
