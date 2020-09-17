package src

import (
	"archive/tar"
	"compress/gzip"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	logger "github.com/sirupsen/logrus"
	"golang.org/x/net/html/charset"
)

var (
	err error
	// 日期转换参数
	yyyyMMdd = "20060102"
	// 交易日历
	tradingDays = sort.StringSlice{}
	// 读取xml tar.gz file path
	xmlFilePath = "/mnt/future_xml"
	// 保存csv文件夹
	csvPath = "/mnt/future_tick_csv_gz"
)

// 初始化
func init() {
	// 环境变量读取
	if tmp := os.Getenv("xmlFilePath"); tmp != "" {
		xmlFilePath = tmp
	}
	if tmp := os.Getenv("csvPath"); tmp != "" {
		csvPath = tmp
	}

	// 日志初始化
	LogInit()
	// 创建csv文件路径
	if _, err := os.Stat(csvPath); err != nil {
		os.Mkdir(csvPath, 0777)
		os.Chmod(csvPath, 0777)
	}
	// 读取交易日历
	readCalendar()
}

func checkErr(err error) {
	if err != nil {
		logger.Panic(err)
	}
}

func readCalendar() {
	// 取交易日历
	cal, err := os.Open("calendar.csv")
	defer cal.Close()
	if err != nil {
		logger.Error(err)
	}
	reader := csv.NewReader(cal)
	lines, err := reader.ReadAll()
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		if line[1] == "true" {
			tradingDays = append(tradingDays, line[0])
		}
	}
	sort.Sort(tradingDays)
}

// Run 运行
func Run(startDay string) {
	if startDay == "" {
		// csv 文件最大文件名
		files, _ := ioutil.ReadDir(csvPath)
		days := []string{}
		for _, f := range files {
			if f.IsDir() {
				continue
			} else {
				days = append(days, strings.Split(f.Name(), ".")[0])
			}
		}
		// 排序取最大完成日期
		if len(days) > 0 {
			ss := sort.StringSlice(days)
			sort.Sort(ss)
			startDay = ss[len(ss)-1]
		}
	} else { // 取指定日期前一交易日,以便程序执行时包括指定的日期
		startIdx := 0
		for i := 0; i < len(tradingDays); i++ {
			if tradingDays[i] == startDay {
				startIdx = i - 1
				break
			}
		}
		startDay = tradingDays[startIdx]
	}
	// xml文件列表
	xmlFiles := sort.StringSlice{}
	files, _ := ioutil.ReadDir(xmlFilePath)
	for _, f := range files {
		if !f.IsDir() {
			name := strings.Split(f.Name(), ".")[0]
			if name > startDay { // >=程序启动时会重新处理最后一天的数据
				xmlFiles = append(xmlFiles, name)
			}
		}
	}
	// xmlFiles = sort.StringSlice(xmlFiles)
	sort.Sort(xmlFiles)
	// for _, tradingDay := range xmlFiles {
	// 	logger.Info(tradingDay, " start...")
	// 	err := XMLToTickData(tradingDay)
	// 	if err != nil {
	// 		logger.Panic(tradingDay, " Error:", err)
	// 	} else {
	// 		logger.Info(tradingDay, " finished.")
	// 	}
	// }
	// 使用chan 控制协程总数
	var waitGroup sync.WaitGroup
	chDay := make(chan string, runtime.NumCPU())
	for _, tradingDay := range xmlFiles {
		chDay <- tradingDay
		waitGroup.Add(1)
		go func(d string) {
			logger.Info(d, " start...")
			err := XMLToTickData(d)
			if err != nil {
				logger.Panic(d, " Error:", err)
			} else {
				logger.Info(d, " finished.")
			}
			<-chDay
			waitGroup.Done()
		}(tradingDay)
	}
	waitGroup.Wait()
	close(chDay)

	// 取下一交易日
	latestDay := startDay
	if len(xmlFiles) > 0 {
		latestDay = xmlFiles[len(xmlFiles)-1]
	}
	latestIdx := 0
	for i := 0; i < len(tradingDays); i++ {
		if tradingDays[i] > latestDay {
			latestIdx = i
			break
		}
	}
	latestDay = tradingDays[latestIdx]
	logger.Info(latestDay, " waiting...")
	for {
		_, err := os.Stat(path.Join(xmlFilePath, latestDay+".tar.gz"))
		// 文件存在
		if err == nil {
			logger.Info(latestDay, " start...")
			XMLToTickData(latestDay)
			latestIdx++
			latestDay = tradingDays[latestIdx]
			logger.Info(latestDay, " waiting...")
		} else { // 从sftp读取
			// export xmlSftp=192.168.111.191/22/root/123456
			if tmp := os.Getenv("xmlSftp"); tmp != "" {
				ss := strings.Split(tmp, "/")
				host, user, pwd := ss[0], ss[2], ss[3]
				port, _ := strconv.Atoi(ss[1])
				//sftp, err := NewHfSftp("192.168.111.191", 22, "root", "123456")
				sftp, err := NewHfSftp(host, port, user, pwd)
				defer sftp.Close()
				checkErr(err)
				srcFile, err := sftp.GetFile(path.Join(os.Getenv("xmlSftpPath"), latestDay+".tar.gz"))
				defer srcFile.Close()
				if err == nil {
					logger.Info(latestDay, " reading...")
					dstFile, err := os.OpenFile(path.Join(xmlFilePath, latestDay+".tar.gz"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
					defer dstFile.Close()
					checkErr(err)
					buf := make([]byte, 1024*1024*100)
					for {
						n, _ := srcFile.Read(buf)
						if n == 0 {
							logger.Info(latestDay, " write finish.")
							logger.Info(latestDay, " to csv start...")
							XMLToTickData(latestDay)
							latestIdx++
							latestDay = tradingDays[latestIdx]
							logger.Info(latestDay, " waiting...")
							break
						}
						dstFile.Write(buf[0:n])
					}
				} else { // 文件不存在
					// logger.Error(err)
					time.Sleep(time.Minute * 10)
				}
			} else { // 未配置 Sftp
				time.Sleep(time.Minute * 10)
			}
		}
	}
}

// XMLToTickData 处理某一天的数据
// tradingDay 交易日
func XMLToTickData(tradingDay string) (err error) {
	// xml => csv
	f, err := os.OpenFile(path.Join(xmlFilePath, tradingDay+".tar.gz"), os.O_RDONLY, os.ModePerm)
	checkErr(err)
	defer f.Close()

	_, err = f.Seek(0, 0) // 切换到文件开始,否则err==EOF
	gr, err := gzip.NewReader(f)
	defer gr.Close()
	if err != nil {
		return err
	}

	tr := tar.NewReader(gr)
	// 解压tar中的所有文件(其实只有一个)
	_, _ = tr.Next()
	// 包中的 marketdata.xml 解析成tick并入库
	decoder := xml.NewDecoder(tr)
	// 处理汉字编码
	decoder.CharsetReader = func(c string, i io.Reader) (io.Reader, error) {
		return charset.NewReaderLabel(strings.TrimSpace(c), i)
	}
	lineToTick(decoder, tradingDay)

	// 结束也会产生error
	if err == io.EOF {
		err = nil
	}
	return nil
}

func lineToTick(decoder *xml.Decoder, tradingDay string) {
	// 处理actionday
	var actionDay, actionNextDay string

	if len(tradingDays) == 1 {
		actionDay = tradingDay
		actionNextDay = tradingDay
	} else {
		idx := -1
		for i := 0; i < len(tradingDays); i++ {
			if tradingDays[i] == tradingDay {
				idx = i
				break
			}
		}
		actionDay = tradingDays[idx-1]
		t, _ := time.Parse(yyyyMMdd, actionDay)
		actionNextDay = t.AddDate(0, 0, 1).Format(yyyyMMdd)
	}

	var cnt = 0
	gz, err := os.OpenFile(path.Join(csvPath, tradingDay+".gz"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm) // os.O_TRUNC覆盖
	checkErr(err)
	defer gz.Close()
	csvGz := gzip.NewWriter(gz)
	defer csvGz.Close()

	csvGz.Write([]byte("TradingDay,InstrumentID,UpdateTime,UpdateMillisec,ActionDay,LowerLimitPrice,UpperLimitPrice,BidPrice1,AskPrice1,AskVolume1,BidVolume1,LastPrice,Volume,OpenInterest,Turnover,AveragePrice\n"))
	// 无法解决 expect EOF 的错误
	for t, err := decoder.Token(); err == nil; t, err = decoder.Token() {
		if t == nil {
			logger.Error("get token error!")
			break
		}

		switch start := t.(type) {
		case xml.StartElement:
			// ...and its name is "page" NtfDepthMarketDataPackage
			se := start.Copy()
			ee := se.End()
			if se.Name.Local == "NtfDepthMarketDataPackage" {
				p := NtfDepthMarketDataPackage{}
				if err = decoder.DecodeElement(&p, &se); err != nil {
					logger.Error(se)
					fmt.Printf("%v", ee)
					logger.Panic(err) // 遇到错误返回 报错:unexpect EOF
				} else {
					// 过虑脏数据
					if p.MarketDataLastMatchField.Volume == 0 || p.MarketDataLastMatchField.LastPrice == 0 || p.MarketDataBestPriceField.AskPrice1 == 0 || p.MarketDataBestPriceField.BidPrice1 == 0 {
						continue
					}
					// 处理actionDay
					if hour, err := strconv.Atoi(p.MarketDataUpdateTimeField.UpdateTime[0:2]); err == nil {
						if hour >= 20 {
							p.MarketDataUpdateTimeField.ActionDay = actionDay
						} else if hour < 4 {
							p.MarketDataUpdateTimeField.ActionDay = actionNextDay
						} else {
							p.MarketDataUpdateTimeField.ActionDay = tradingDay
						}
					} else {
						logger.Panic(err)
					}
					p.MarketDataBaseField.TradingDay = tradingDay
					cnt++
					csvGz.Write([]byte(fmt.Sprintf("%s,%s,%s,%d,%s,%.4f,%.4f,%.4f,%.4f,%d,%d,%.4f,%d,%.4f,%.4f,%.4f\n", p.MarketDataBaseField.TradingDay, p.MarketDataUpdateTimeField.InstrumentID, p.MarketDataUpdateTimeField.UpdateTime, p.MarketDataUpdateTimeField.UpdateMillisec, p.MarketDataUpdateTimeField.ActionDay, p.MarketDataStaticField.LowerLimitPrice, p.MarketDataStaticField.UpperLimitPrice, p.MarketDataBestPriceField.BidPrice1, p.MarketDataBestPriceField.AskPrice1, p.MarketDataBestPriceField.AskVolume1, p.MarketDataBestPriceField.BidVolume1, p.MarketDataLastMatchField.LastPrice, p.MarketDataLastMatchField.Volume, p.MarketDataLastMatchField.OpenInterest, p.MarketDataLastMatchField.Turnover, p.MarketDataAveragePriceField.AveragePrice)))
				}
			}
		}
	}
	// 完成后改名,避免下一步操作读到未完成的数据
	os.Rename(path.Join(csvPath, tradingDay+".gz"), path.Join(csvPath, tradingDay+".csv.gz"))
	logger.Info(tradingDay, ":", cnt)
}
