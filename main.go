package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Config struct {
	ConstituentWord []string            `json:"constituentWord"`
	EndWord         []string            `json:"endWord"`
	MutexGroups     map[string][]string `json:"mutexGroups"`
	BaseDomains     []string            `json:"baseDomains"`
	Concurrency     int                 `json:"concurrency"`
}

type URLStatus struct {
	URL          string
	HeaderStatus int
	BodyStatus   int
}

func loadConfig(path string) (Config, error) {
	var config Config
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	return config, err
}

func isMutex(word, requisite string, mutexGroups map[string][]string) bool {
	for k, v := range mutexGroups {
		if k == word && contains(v, requisite) {
			return true
		}
		if k == requisite && contains(v, word) {
			return true
		}
	}
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

type ResponseJSON struct {
	Code int `json:"code"`
}

// generateCombinations 生成所有可能的组合，不使用并发
func generateCombinations(constituents []string) [][]string {
	var result [][]string
	for i := 0; i < (1 << len(constituents)); i++ {
		var combination []string
		for j, word := range constituents {
			if i&(1<<j) > 0 {
				combination = append(combination, word)
			}
		}
		result = append(result, combination)
	}
	return result
}

func hasMutex(combination []string, mutexGroups map[string][]string) bool {
	for i, word := range combination {
		for j, otherWord := range combination {
			if i != j && isMutex(word, otherWord, mutexGroups) {
				return true
			}
		}
	}
	return false
}

// 测试连通性，超时设置为5秒
func canPing(domain string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	_, err := client.Get(domain)
	return err == nil
}

func main() {
	rand.Seed(time.Now().UnixNano())
	client := &http.Client{Timeout: 10 * time.Second}
	logFile, err := os.Create("EndpointJoinerの运行日志.txt")
	mw := io.MultiWriter(os.Stdout, logFile)
	log := func(format string, a ...interface{}) {
		fmt.Fprintf(mw, format+"\n", a...)
	}
	if err != nil {
		log("xxx 创建日志文件失败: %v\n", err)
		return
	}
	defer logFile.Close()

	log("正在加载配置...\n")
	config, err := loadConfig("config.json")
	if err != nil {
		log("xxx 加载配置失败: %v\n", err)
		return
	}

	log("√√√ 设置并发数为 %d", config.Concurrency)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, config.Concurrency)

	log("\n正在生成端点组合...\n")
	combinations := generateCombinations(config.ConstituentWord)
	log("√√√ 端点组合生成完成，共生成 %d 个组合\n", len(combinations))

	codeCounts := make(map[int]int)
	headerCodeCounts := make(map[int]int)
	urlStatuses := []URLStatus{}
	urlStatusesMutex := &sync.Mutex{}

	var validRequests int
	log("正在执行HTTP请求...\n")
	for _, domain := range config.BaseDomains {
		log("正在检查域名 %s 是否有效...\n", domain)
		if !canPing(domain) {
			log("xxx 域名 %s 无效，跳过...\n", domain)
			continue
		}

		for _, end := range config.EndWord {
			for _, combination := range combinations {
				if hasMutex(combination, config.MutexGroups) {
					continue
				}

				semaphore <- struct{}{}
				wg.Add(1)
				go func(domain, end string, combination []string) {
					defer wg.Done()
					defer func() { <-semaphore }()

					url := domain + strings.Join(combination, "/")
					if !strings.HasPrefix(end, "/") {
						url += "/"
					}
					url += end

					resp, err := client.Get(url)
					if err != nil {
						log("xxx 请求 %s 失败: %v\n", url, err)
						return
					}
					defer resp.Body.Close()

					headerStatus := resp.StatusCode
					var responseJSON ResponseJSON
					if err := json.NewDecoder(resp.Body).Decode(&responseJSON); err != nil {
						log("xxx 解析 %s 响应失败: %v\n", url, err)
						return
					}

					log("%s\n响应: %d, 响应体code: %d\n", url, headerStatus, responseJSON.Code)

					urlStatusesMutex.Lock()
					codeCounts[responseJSON.Code]++
					headerCodeCounts[headerStatus]++

					validRequests++
					if responseJSON.Code != 404 {
						urlStatuses = append(urlStatuses, URLStatus{URL: url, HeaderStatus: headerStatus, BodyStatus: responseJSON.Code})
					}
					urlStatusesMutex.Unlock()
				}(domain, end, combination)
			}
		}
	}

	wg.Wait()
	log("√√√ 任务完成\n")

	if len(headerCodeCounts) > 0 {
		for code, count := range headerCodeCounts {
			log("结果统计:")
			log("响应头状态码[%d]: %d次", code, count)
		}
	}
	if len(codeCounts) > 0 {
		for code, count := range codeCounts {
			log("响应体code[%d]: %d次", code, count)
		}
	}

	if len(urlStatuses) > 0 {
		log("其他状态码的详细信息:\n")
		for _, status := range urlStatuses {
			log("请求: %s, 响应头状态码: %d, 响应体中的code状态码: %d\n", status.URL, status.HeaderStatus, status.BodyStatus)
		}
	}
}
