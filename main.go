package main

import (
	"flag"
	"fmt"
	"src/src"
)

var (
	singleDay = ""
	startDay  = ""
)

func init() {
	flag.StringVar(&singleDay, "s", "", "处理指定某一天的数据")
	flag.StringVar(&startDay, "m", "", "从指定日期开始处理数据, 默认为空, tick.csv 文件下一日处理")
}

func main() {
	flag.Parse()
	if singleDay != "" {
		err := src.XMLToTickData(singleDay)
		if err != nil {
			fmt.Print(err)
		}
	} else {
		src.Run(startDay)
	}
}
