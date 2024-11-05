package cmd

import (
	"fmt"
	"github.com/apernet/hysteria/core/v2/client"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"net"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"syscall"
	"time"
)

// Proxy 代表clash配置部分
type Proxy struct {
	Name           string `yaml:"name"`
	Server         string `yaml:"server"`
	Port           int    `yaml:"port"`
	Ports          string `yaml:"ports"`
	MPort          string `yaml:"mport"`
	UDP            bool   `yaml:"udp"`
	SkipCertVerify bool   `yaml:"skip-cert-verify"`
	SNI            string `yaml:"sni"`
	Type           string `yaml:"type"`
	Password       string `yaml:"password"`
	Obfs           string `yaml:"obfs"`
	ObfsPassword   string `yaml:"obfs-password"`
	PingLatency    time.Duration
}

type MyClientConfig struct {
	StartPort        int      `json:"start_port"`
	Count            int      `json:"count"`
	ClashConfigFiles []string `json:"clash_config_files"`
}

// clashConfig 是Clash配置结构体，包含多个代理配置
type clashConfig struct {
	Proxies []Proxy `yaml:"proxies"`
}

// h2clientConfig 将Clash配置转换为Hysteria2客户端配置
func (c *clashConfig) h2clientConfig() []*clientConfig {
	var ret []*clientConfig
	for _, proxy := range c.Proxies {
		cc := clientConfig{
			Server: fmt.Sprintf("%s:%d", proxy.Server, proxy.Port),
			Auth:   proxy.Password,
		}
		if proxy.Obfs != "" && proxy.ObfsPassword != "" {
			cc.Obfs = clientConfigObfs{
				proxy.Obfs,
				clientConfigObfsSalamander{
					Password: proxy.ObfsPassword,
				},
			}
		}
		ret = append(ret, &cc)
	}
	return ret
}

// parseConfig 解析配置文件，返回解析后的Clash配置
func parseConfig(files []string, max int) (*clashConfig, error) {
	serverMap := make(map[string]Proxy)
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", file, err)
		}
		var config clashConfig
		if err = yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal YAML in %s: %w", file, err)
		}
		for _, proxy := range config.Proxies {
			if proxy.Type != "hysteria2" {
				continue
			}
			if _, ok := serverMap[proxy.Server]; ok {
				continue
			}
			serverMap[proxy.Server] = proxy
		}
	}

	// 按照延迟排序，取前 max 个
	proxies := make([]Proxy, 0, len(serverMap))
	for _, proxy := range serverMap {
		// 获取 proxy.Server 的“延迟”值
		latency, err := getPingLatency(proxy.Server, proxy.Port)
		if err != nil {
			logger.Warn("failed to measure latency for server", zap.String("server", proxy.Server), zap.Error(err))
			continue
		}
		proxy.PingLatency = latency
		proxies = append(proxies, proxy)
	}
	sort.Slice(proxies, func(i, j int) bool {
		return proxies[i].PingLatency < proxies[j].PingLatency
	})
	if len(proxies) > max {
		proxies = proxies[:max]
	}

	return &clashConfig{proxies}, nil
}

// getPingLatency 测量服务器的延迟
func getPingLatency(server string, port int) (time.Duration, error) {
	address := fmt.Sprintf("%s:%d", server, port)
	start := time.Now()
	conn, err := net.DialTimeout("udp", address, time.Second*5)
	if err != nil {
		return 0, fmt.Errorf("failed to connect to server %s: %w", address, err)
	}
	defer conn.Close()
	return time.Since(start), nil // 返回连接所花费的时间
}

// MyClientRun 是客户端运行的主函数
func MyClientRun(c MyClientConfig) {
	initLogger()
	clashConfig, err := parseConfig(c.ClashConfigFiles, c.Count)
	if err != nil {
		logger.Fatal("failed to parse config", zap.Error(err))
	}
	configs := clashConfig.h2clientConfig()
	clients := map[string]client.Client{}
	var runner clientModeRunner

	// 初始化所有客户端
	err = initClients(configs, &runner, clients, c.StartPort)
	if err != nil {
		logger.Fatal("failed to initialize clients", zap.Error(err))
	}
	// 运行客户端模式
	runClientMode(&runner, clients)
}

// initClients 初始化客户端并将其添加到 runner 中
func initClients(configs []*clientConfig, runner *clientModeRunner, clients map[string]client.Client, startPort int) error {
	for i, config := range configs {
		hc := httpConfig{
			Listen: fmt.Sprintf("127.0.0.1:%d", i+startPort),
		}
		logger.Info("config:", zap.String("addr", hc.Listen), zap.String("Server", config.Server))
		cli, err := client.NewReconnectableClient(
			config.Config,
			func(c client.Client, info *client.HandshakeInfo, count int) {
				logger.Info("connected to server",
					zap.Bool("udpEnabled", info.UDPEnabled),
					zap.Uint64("tx", info.Tx),
					zap.Int("count", count))
			},
			true,
		)
		if err != nil {
			return fmt.Errorf("failed to initialize client for server %s: %w", config.Server, err)
		}

		clients[config.Server] = cli
		runner.Add("HTTP proxy server"+strconv.Itoa(i), func() error {
			return clientHTTP(hc, cli)
		})
	}
	return nil
}

// runClientMode 运行客户端模式并监听结果
func runClientMode(runner *clientModeRunner, clients map[string]client.Client) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	runnerChan := make(chan clientModeRunnerResult, 1)
	go func() {
		runnerChan <- runner.Run()
	}()

	select {
	case r := <-runnerChan:
		if r.OK {
			logger.Info(r.Msg)
		} else {
			for cli := range clients {
				clients[cli].Close()
			}
			if r.Err != nil {
				logger.Fatal(r.Msg, zap.Error(r.Err))
			} else {
				logger.Fatal(r.Msg)
			}
		}
	case <-signalChan:
		logger.Info("received signal, shutting down gracefully")
		for key, cli := range clients {
			cli.Close()
			logger.Info(key + ":Closed")
		}
		signal.Stop(signalChan)
		logger.Info("signal:Closed")
	}

}
