package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/helloworlde/miwifi-exporter/config"
)

// fetchTimeout 是单次请求路由器的超时时间。
// 旧代码用零值 http.Client{}（无超时），请求挂起会阻塞整个采集周期，这里加上超时。
const fetchTimeout = 15 * time.Second

// apiURL 拼接小米路由器带 stok 鉴权的 API 地址。
func apiURL(path string) string {
	return fmt.Sprintf("http://%s/cgi-bin/luci/;stok=%s/%s",
		config.Configs.IP, config.Token.Token, path)
}

// fetchAPI 请求路由器 API 并把 JSON 反序列化到 target。
//
// 相比原先每个 Get 函数各写一遍的 client.Get + Unmarshal + 失败重试逻辑，
// 这里集中处理并修复了几个已知问题：
//   - 旧代码 Unmarshal 失败时的 count 是函数局部变量且不在循环里，count >= 5 永远不会触发；
//     GetMiSystemStatus 还会无限递归调用自身。
//   - 旧代码在 token 失效时 sleep 1 分钟，会让整个 Prometheus scrape 超时。
//
// 现在：Unmarshal 失败（通常是 token 失效）→ 调用 config.GetConfig() 重新登录 → 重试 1 次。
// 仍然失败则返回 error，由调用方记录日志后跳过本次采集，下一次 scrape 自然恢复，
// 而不是 os.Exit(1) 让整个进程退出。
func fetchAPI(path string, target interface{}) error {
	if err := doGet(apiURL(path), target); err == nil {
		return nil
	}

	// 第一次失败，很可能是 token 过期，重新登录后再试一次。
	config.GetConfig()
	if err := doGet(apiURL(path), target); err != nil {
		return fmt.Errorf("token 失效或路由器鉴权错误: %w", err)
	}
	return nil
}

// doGet 执行一次 GET 并反序列化。
func doGet(url string, target interface{}) error {
	client := &http.Client{Timeout: fetchTimeout}
	res, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("请求路由器错误（可能路由器掉线或宕机）: %w", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}
	return nil
}
