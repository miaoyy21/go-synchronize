package base

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func Init(dir string) error {
	// 获取配置文件
	bytes, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		return err
	}

	// 解析
	if err := json.Unmarshal(bytes, &Config); err != nil {
		return err
	}

	Config.Dir = dir
	return nil
}

func CompareMap(latest map[string]string, present map[string]string) (map[string]string, map[string]string, map[string]string) {
	added, changed, removed := make(map[string]string), make(map[string]string), make(map[string]string)

	// 增加
	for key, value := range present {
		xvalue, ok := latest[key]
		if !ok {
			added[key] = value
		} else if xvalue != value {
			// 更新
			changed[key] = value
		}
	}

	// 移除
	for key, value := range latest {
		if _, ok := present[key]; !ok {
			removed[key] = value
		}
	}

	return added, changed, removed
}
