# ![Hysteria 2](logo.svg)

# 用hysteria2做ip池
+ 起因爬虫做了ip的并发限制，大量爬取之后ip被封，所以想用hysteria2做ip池.
+ 为什么用hysteria2也是因为目前来看hysteria2用的比ss多
+ hysteria2/clash/ss 这些本身编译好的程序都是不支持 访问同一个网站的时候使用不同ip的.唯一方案是开启多个客户端,但是很麻烦.
+ ip池本身使用的是clash的配置文件
+ 这种pr官方肯定不会接受... 放一下魔改了的源码
### 结果
```bash
2024-11-04T00:30:38+08:00       INFO    HTTP proxy server listening     {"addr": "127.0.0.1:8080"}
2024-11-04T00:30:38+08:00       INFO    HTTP proxy server listening     {"addr": "127.0.0.1:8088"}
2024-11-04T00:30:38+08:00       INFO    HTTP proxy server listening     {"addr": "127.0.0.1:8085"}
2024-11-04T00:30:38+08:00       INFO    HTTP proxy server listening     {"addr": "127.0.0.1:8086"}
2024-11-04T00:30:38+08:00       INFO    HTTP proxy server listening     {"addr": "127.0.0.1:8089"}
2024-11-04T00:30:38+08:00       INFO    HTTP proxy server listening     {"addr": "127.0.0.1:8081"}
2024-11-04T00:30:38+08:00       INFO    HTTP proxy server listening     {"addr": "127.0.0.1:8084"}
2024-11-04T00:30:38+08:00       INFO    HTTP proxy server listening     {"addr": "127.0.0.1:8082"}
2024-11-04T00:30:38+08:00       INFO    HTTP proxy server listening     {"addr": "127.0.0.1:8087"}
2024-11-04T00:30:38+08:00       INFO    HTTP proxy server listening     {"addr": "127.0.0.1:8083"}
```

### 修改了两个文件: `app/cmd/myclient.go`和`app/main.go`

配置文件如下`myclient.json`:  
+ start_port 从那个端口开始代理
+ count 代理的几个端口
+ clash_config_files clash的配置文件,数组可以填多个,会自动读取所有hysteria2的配置,并通过延迟决定前count个
```json
{
  "start_port": 7000,
  "count": 10,
  "clash_config_files": [
    "/Users/parapeng/Library/Application Support/io.github.clash-verge-rev.clash-verge-rev/profiles/ROO5OxI3HLEr.yaml"
  ]
}
```
### 使用:

下载自己的release,和myclient.json,配置自己的clash_config_files文件即可

### 线程池
线程池可以参考[poolhttp](https://github.com/pzx521521/pixelcut/blob/master/poolhttp.go)
以及[poolhttp_test](https://github.com/pzx521521/pixelcut/blob/master/poolhttp_test.go)
示例结果如下:
```bash
2024/11/05 16:41:56 resp.Body: {"ip":"83.147.17.189","country":"GB","country_name":"United Kingdom","region_code":"ENG","in_eu":true,"continent":"EU"}
2024/11/05 16:41:56 resp.Body: {"ip":"83.147.17.189","country":"GB","country_name":"United Kingdom","region_code":"ENG","in_eu":true,"continent":"EU"}
2024/11/05 16:41:57 resp.Body: {"ip":"61.224.133.143","country":"TW","country_name":"Taiwan","region_code":"TXG","in_eu":false,"continent":"AS"}
2024/11/05 16:41:57 resp.Body: {"ip":"61.224.133.143","country":"TW","country_name":"Taiwan","region_code":"TXG","in_eu":false,"continent":"AS"}
2024/11/05 16:41:57 resp.Body: {"ip":"107.189.29.215","country":"LU","country_name":"Luxembourg","region_code":"LU","in_eu":true,"continent":"EU"}
2024/11/05 16:41:57 resp.Body: {"ip":"107.189.29.215","country":"LU","country_name":"Luxembourg","region_code":"LU","in_eu":true,"continent":"EU"}
2024/11/05 16:41:57 resp.Body: {"ip":"184.174.96.224","country":"US","country_name":"United States","region_code":"DE","in_eu":false,"continent":"NA"}
2024/11/05 16:41:57 resp.Body: {"ip":"184.174.96.224","country":"US","country_name":"United States","region_code":"DE","in_eu":false,"continent":"NA"}
2024/11/05 16:41:57 resp.Body: {"ip":"209.200.246.141","country":"CA","country_name":"Canada","region_code":"ON","in_eu":false,"continent":"NA"}
2024/11/05 16:41:57 resp.Body: {"ip":"209.200.246.141","country":"CA","country_name":"Canada","region_code":"ON","in_eu":false,"continent":"NA"}
2024/11/05 16:41:57 resp.Body: {"ip":"87.121.61.171","country":"FR","country_name":"France","region_code":"GES","in_eu":true,"continent":"EU"}
2024/11/05 16:41:57 resp.Body: {"ip":"83.147.17.189","country":"GB","country_name":"United Kingdom","region_code":"ENG","in_eu":true,"continent":"EU"}
2024/11/05 16:41:57 resp.Body: {"ip":"83.147.17.189","country":"GB","country_name":"United Kingdom","region_code":"ENG","in_eu":true,"continent":"EU"}
2024/11/05 16:41:58 resp.Body: {"ip":"61.224.133.143","country":"TW","country_name":"Taiwan","region_code":"TXG","in_eu":false,"continent":"AS"}
2024/11/05 16:41:58 resp.Body: {"ip":"61.224.133.143","country":"TW","country_name":"Taiwan","region_code":"TXG","in_eu":false,"continent":"AS"}
2024/11/05 16:41:58 resp.Body: {"ip":"107.189.29.215","country":"LU","country_name":"Luxembourg","region_code":"LU","in_eu":true,"continent":"EU"}
2024/11/05 16:41:58 resp.Body: {"ip":"107.189.29.215","country":"LU","country_name":"Luxembourg","region_code":"LU","in_eu":true,"continent":"EU"}
2024/11/05 16:41:58 resp.Body: {"ip":"184.174.96.224","country":"US","country_name":"United States","region_code":"DE","in_eu":false,"continent":"NA"}
2024/11/05 16:41:58 resp.Body: {"ip":"184.174.96.224","country":"US","country_name":"United States","region_code":"DE","in_eu":false,"continent":"NA"}
2024/11/05 16:41:58 resp.Body: {"ip":"87.121.61.171","country":"FR","country_name":"France","region_code":"GES","in_eu":true,"continent":"EU"}

```

成果是一个[壁纸网站](https://paral.us.kg/):