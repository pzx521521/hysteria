package main

import (
	"encoding/json"
	"github.com/apernet/hysteria/app/v2/cmd"
	"log"
	"os"
)

func main() {
	// 添加配置文件的搜索路径
	data, err := os.ReadFile("./myclient.json")
	if err != nil {
		log.Fatal(err)
	}
	var myClientConfig cmd.MyClientConfig
	err = json.Unmarshal(data, &myClientConfig)
	if err != nil {
		log.Fatal(err)
	}
	cmd.MyClientRun(myClientConfig)
}
