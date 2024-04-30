package comfyUIclient

import (
	"encoding/json"
	"strings"
	"time"
)

// TraverseAndModifySeed Reset all seed values for the workflow
func TraverseAndModifySeed(workflow string) string {
	var data map[string]interface{}
	json.Unmarshal([]byte(workflow), &data)
	traverseAndModifySeed(data)
	res, _ := json.Marshal(data)
	return string(res)
}

// TraverseAndModifyJSON Replace the values ​​of nodes in the workflow.
//
//	{
//		"3": {
//		"inputs": {
//			"seed": 4702551705662,
//		}
//	}
//
// For the seed node above, the keys should be ["3", "inputs", "seed"].
func TraverseAndModifyJSON(workflow string, keys []string, newValue interface{}) string {
	var data map[string]interface{}
	json.Unmarshal([]byte(workflow), &data)
	traverseAndModifyJSON(data, keys, newValue)
	res, _ := json.Marshal(data)
	return string(res)
}

// BatchTraverseAndModifyJSON Batch replace the values of nodes in the workflow.
//
//	{
//		"3": {
//		"inputs": {
//			"seed": 4702551705662,
//		}
//	}
//
//	For the seed node above, the nodeKey should be "3-inputs-seed", Val shoule be integer.
func BatchTraverseAndModifyJSON(workflow string, nodeKeyVal map[string]interface{}, seed bool) string {
	var data map[string]interface{}
	json.Unmarshal([]byte(workflow), &data)

	if seed {
		traverseAndModifySeed(data)
	}

	for nodeKey, val := range nodeKeyVal {
		keys := strings.Split(nodeKey, "-")
		traverseAndModifyJSON(data, keys, val)
	}

	res, _ := json.Marshal(data)
	return string(res)
}

func traverseAndModifySeed(data interface{}) {
	switch value := data.(type) {
	case map[string]interface{}:
		for k, v := range value {
			if k == "seed" {
				value[k] = time.Now().UnixMicro()
			}
			traverseAndModifySeed(v)
		}
	case []interface{}:
		for _, val := range value {
			traverseAndModifySeed(val)
		}
	}
}

func traverseAndModifyJSON(data interface{}, keys []string, newValue interface{}) {
	if len(keys) == 0 {
		return
	}
	switch value := data.(type) {
	case map[string]interface{}:
		if len(keys) == 1 {
			key := keys[0]

			if _, ok := value[key]; ok {
				value[key] = newValue
			}
		} else {
			if val, ok := value[keys[0]]; ok {
				traverseAndModifyJSON(val, keys[1:], newValue)
			}
		}
	case []interface{}:
		for _, val := range value {
			traverseAndModifyJSON(val, keys, newValue)
		}
	}
}
