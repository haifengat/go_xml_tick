## go xml tick
golang 实现读取xml文件,并转换为csv文件.

## 环境变量
* xmlFilePath
    * xml.tar.gz文件路径
* csvPath
    * 保存tick数据的csv文件路径
* xml
    * 文件所在的sftp配置,不配置则不读取
    * 格式: ip/port/user/password

## 格式
* 文件
> gzip压缩
* float
>  采用.4f格式
* 标题
> TradingDay,InstrumentID,UpdateTime,UpdateMillisec,ActionDay,LowerLimitPrice,UpperLimitPrice,BidPrice1,AskPrice1,AskVolume1,BidVolume1,LastPrice,Volume,OpenInterest,Turnover,AveragePrice

## Dockerfile
go build -o bin/xml_tick main.go
```dockerfile
FROM haifengat/ctp_real_md

COPY bin/xml_tick /home
RUN chmod a+x /home/xml_tick
ENTRYPOINT ["/home/xml_tick"]
```

## build
```bash
docker build -t haifengat/go_xml_tick .
# 通过github git push触发 hub.docker自动build
docker pull haifengat/go_xml_tick && docker tag haifengat/go_xml_tick haifengat/go_xml_tick:`date +%Y%m%d` && docker push haifengat/go_xml_tick:`date +%Y%m%d`
```

### 启动
```bash
docker-compose --compatibility up -d
```

## docker-compose.yml
```yml
version: "3"
services:
go_xml_tick:
    image: haifengat/go_xml_tick
    container_name: go_xml_tick
    restart: always
    environment:
        - TZ=Asia/Shanghai
        - xmlFilePath=/home/xml_path
        - csvPath=/home/csv_path
    volumes: 
        - /mnt/future_xml:/home/xml_path
        - /mnt/future_tick_csv_gz:/home/csv_path
    deploy:
        resources:
            limits:
                memory: 6G
            reservations:
                memory: 200M
```