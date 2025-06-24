package utils

import "speedtest-gd/global"

func GetString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
func GetInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	return 0
}
func GetInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key].(float64); ok {
		return int64(v)
	}
	if v, ok := m[key].(int64); ok {
		return v
	}
	return 0
}

func MergeUnique(agents ...[]global.ApacheAgent) []global.ApacheAgent {
	seen := make(map[string]struct{})
	result := make([]global.ApacheAgent, 0)
	for _, list := range agents {
		for _, item := range list {
			if _, ok := seen[item.HostIP]; !ok {
				seen[item.HostIP] = struct{}{}
				result = append(result, item)
			}
		}
	}
	return result
}
