## go xml tick
golang 实现读取xml文件,并转换为csv文件.

## 环境变量
* xmlFilePath = /xml
    * xml.tar.gz文件路径
* csvPath = /csv
    * 保存tick数据的csv文件路径
* xmlSftp
    * 文件所在的sftp配置,不配置则不读取
    * 格式: ip/port/user/password
* xmlSftpPath
    * sftp登录后取xml.tag.gz文件的路径

## 格式
* 文件
> gzip压缩
* float
>  采用.4f格式
* 标题
> TradingDay,InstrumentID,UpdateTime,UpdateMillisec,ActionDay,LowerLimitPrice,UpperLimitPrice,BidPrice1,AskPrice1,AskVolume1,BidVolume1,LastPrice,Volume,OpenInterest,Turnover,AveragePrice

## Dockerfile
```dockerfile
FROM golang:1.14-alpine3.11 AS builder

ENV GOPROXY https://goproxy.cn

WORKDIR /build
COPY go.mod .
COPY go.sum .

# 新增用户
RUN adduser -u 10001 -D app-runner
# 编译
COPY . .
COPY ./src ./src
RUN go mod download; \
    CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -o run .;

FROM alpine:3.11 AS final

WORKDIR /app
COPY --from=builder /build/run /app/
# 国内不稳定
RUN wget https://raw.githubusercontent.com/haifengat/ctp_real_md/master/calendar.csv;

#USER app-runner
ENTRYPOINT ["./run"]
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
version: "3.7"
# docker-compose --compatibility up -d
services:
    go_xml_tick:
        image: haifengat/go_xml_tick
        container_name: go_xml_tick
        restart: always
        environment:
            - TZ=Asia/Shanghai
            # xml文件所在的sftp配置,不配置则不读取
            # - xmlSftp=192.168.111.191/22/root/123456
            # - xmlSftpPath=/home/haifeng/data/
        volumes: 
            - /mnt/future_xml:/xml
            - /mnt/future_tick_csv_gz:/csv
        deploy:
            resources:
                limits:
                    memory: 6G
                reservations:
                    memory: 200M
```