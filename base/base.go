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
