package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/levigross/grequests"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Transmission RPC request structure
type TransmissionRequest struct {
	Method    string      `json:"method"`
	Arguments interface{} `json:"arguments,omitempty"`
}

// Transmission RPC response structure
type TransmissionResponse struct {
	Result    string                 `json:"result"`
	Arguments map[string]interface{} `json:"arguments"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	// 提示用户输入第一个 Transmission 服务器地址
	fmt.Print("请输入第一个 Transmission 服务器地址: ")
	transmissionAd, _ := reader.ReadString('\n')
	transmissionAd = strings.TrimSpace(transmissionAd) // 去掉换行符
	transmissionURL := transmissionAd + "/transmission/rpc"

	// 提示用户输入第一个 Transmission 用户名和密码
	fmt.Print("请输入第一个 Transmission 用户名: ")
	username1, _ := reader.ReadString('\n')
	username1 = strings.TrimSpace(username1)

	fmt.Print("请输入第一个 Transmission 密码: ")
	password1, _ := reader.ReadString('\n')
	password1 = strings.TrimSpace(password1)

	// 提示用户输入第二个 Transmission 服务器地址
	fmt.Print("请输入第二个 Transmission 服务器地址: ")
	secondaryTransmissionAd, _ := reader.ReadString('\n')
	secondaryTransmissionAd = strings.TrimSpace(secondaryTransmissionAd)
	secondaryTransmissionURL := secondaryTransmissionAd + "/transmission/rpc"

	// 提示用户输入第二个 Transmission 用户名和密码
	fmt.Print("请输入第二个 Transmission 用户名: ")
	username2, _ := reader.ReadString('\n')
	username2 = strings.TrimSpace(username2)

	fmt.Print("请输入第二个 Transmission 密码: ")
	password2, _ := reader.ReadString('\n')
	password2 = strings.TrimSpace(password2)

	// 提示用户输入种子源目录
	fmt.Print("请输入种子源目录: ")
	sourceTorrentPath, _ := reader.ReadString('\n')
	sourceTorrentPath = strings.TrimSpace(sourceTorrentPath)

	// 提示用户输入目标种子目录
	fmt.Print("请输入目标种子目录: ")
	targetTorrentPath, _ := reader.ReadString('\n')
	targetTorrentPath = strings.TrimSpace(targetTorrentPath)

	// 提示用户输入 passkey
	fmt.Print("请输入 passkey: ")
	passkey, _ := reader.ReadString('\n')
	passkey = strings.TrimSpace(passkey)

	// 检查目录是否存在
	if _, err := os.Stat(sourceTorrentPath); os.IsNotExist(err) {
		log.Fatalf("目录不存在: %v", err)
	}

	// 创建请求体，获取所有种子的所有信息
	req := TransmissionRequest{
		Method: "torrent-get",
		Arguments: map[string]interface{}{
			"fields": []string{"hashString", "trackerStats", "downloadDir"}, // 请求需要的字段
		},
	}

	// 将请求体转换为 JSON
	jsonReq, err := json.Marshal(req)
	if err != nil {
		log.Fatalf("Error encoding request: %v", err)
	}

	// 设置请求选项，包括基本身份认证
	ro := &grequests.RequestOptions{
		JSON:    jsonReq,
		Auth:    []string{username1, password1},
		Headers: map[string]string{"X-Transmission-Session-Id": getSessionID(transmissionURL, username1, password1)},
	}

	// 发送 POST 请求
	resp, err := grequests.Post(transmissionURL, ro)
	if err != nil {
		log.Fatalf("Error making request to Transmission: %v", err)
	}

	// 检查响应状态
	if !resp.Ok {
		log.Fatalf("Transmission API request failed with status: %d", resp.StatusCode)
	}

	// 解析响应
	var transmissionResp TransmissionResponse
	err = json.Unmarshal(resp.Bytes(), &transmissionResp)
	if err != nil {
		log.Fatalf("Error parsing response: %v", err)
	}

	// 计数器，统计符合条件且去重后的种子总数
	matchCount := 0
	seenTorrents := make(map[string]string) // 存储已经处理的 hashString 和对应的 downloadDir

	// 过滤符合条件的种子信息
	for _, torrent := range transmissionResp.Arguments["torrents"].([]interface{}) {
		torrentData := torrent.(map[string]interface{})
		hashString := torrentData["hashString"].(string)
		trackerStats := torrentData["trackerStats"].([]interface{})
		downloadDir := torrentData["downloadDir"].(string)

		// 如果该 hashString 已经处理过，跳过
		if _, seen := seenTorrents[hashString]; seen {
			continue
		}

		// 收集当前 hashString 的所有 trackerStats
		var announceURLs []string
		for _, tracker := range trackerStats {
			trackerMap := tracker.(map[string]interface{})
			announceURL := trackerMap["announce"].(string)

			if containsPasskey(announceURL, passkey) {
				announceURLs = append(announceURLs, announceURL)
			}
		}

		// 打印 hashString 及其所有符合条件的 trackerStats 和 downloadDir
		if len(announceURLs) > 0 {
			fmt.Printf("Torrent hash: %s, Announce URLs: %s, Download Directory: %s\n", hashString, strings.Join(announceURLs, ", "), downloadDir)
			matchCount++                           // 符合条件的种子计数器加一
			seenTorrents[hashString] = downloadDir // 记录 hashString 和 downloadDir

			// 复制种子文件到目标目录
			sourceFile := filepath.Join(sourceTorrentPath, hashString+".torrent")
			destFile := filepath.Join(targetTorrentPath, hashString+".torrent")
			copyFile(sourceFile, destFile)
		}
	}

	// 上传到第二个 Transmission 服务器并设置 downloadDir
	for hashString, downloadDir := range seenTorrents {
		torrentFile := filepath.Join(targetTorrentPath, hashString+".torrent")
		uploadAndSetDownloadDir(secondaryTransmissionURL, username2, password2, torrentFile, downloadDir)
	}

	// 打印符合条件且去重后的种子总数
	fmt.Printf("Total matched torrents (deduplicated): %d\n", matchCount)
}

// 复制文件函数
func copyFile(src, dst string) {
	sourceFile, err := os.Open(src)
	if err != nil {
		log.Fatalf("无法打开源文件: %v", err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		log.Fatalf("无法创建目标文件: %v", err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		log.Fatalf("复制文件失败: %v", err)
	}
	fmt.Printf("文件已复制: %s -> %s\n", src, dst)
}

// 上传种子并设置 downloadDir
func uploadAndSetDownloadDir(url, username, password, torrentFile, downloadDir string) {
	// 读取种子文件
	fileContent, err := os.ReadFile(torrentFile)
	if err != nil {
		log.Fatalf("无法读取种子文件: %v", err)
	}

	// 创建请求体，上传种子
	req := TransmissionRequest{
		Method: "torrent-add",
		Arguments: map[string]interface{}{
			"metainfo":     fileContent, // 将种子文件内容以 base64 编码上传
			"download-dir": downloadDir,
		},
	}

	// 将请求体转换为 JSON
	jsonReq, err := json.Marshal(req)
	if err != nil {
		log.Fatalf("Error encoding request: %v", err)
	}

	// 获取 session ID
	ro := &grequests.RequestOptions{
		JSON:    jsonReq,
		Auth:    []string{username, password},
		Headers: map[string]string{"X-Transmission-Session-Id": getSessionID(url, username, password)},
	}

	// 上传种子
	resp, err := grequests.Post(url, ro)
	if err != nil {
		log.Fatalf("Error uploading torrent: %v", err)
	}

	// 检查响应状态
	if !resp.Ok {
		log.Fatalf("上传种子失败: %d", resp.StatusCode)
	}

	// 上传成功后设置 downloadDir
	var addResp TransmissionResponse
	err = json.Unmarshal(resp.Bytes(), &addResp)
	if err != nil {
		log.Fatalf("Error parsing response: %v", err)
	}
}

// 获取 Transmission 会话 ID
func getSessionID(url, username, password string) string {
	ro := &grequests.RequestOptions{
		Auth: []string{username, password},
	}
	// 发送一个空请求来获取会话 ID
	resp, err := grequests.Post(url, ro)
	if err != nil {
		log.Fatalf("Error getting session ID: %v", err)
	}
	// 从响应头中获取 X-Transmission-Session-Id
	sessionID := resp.Header.Get("X-Transmission-Session-Id")
	if sessionID == "" {
		log.Fatalf("未能获取 X-Transmission-Session-Id")
	}
	return sessionID
}

// 判断 announceURL 是否包含指定的 passkey
func containsPasskey(announceURL string, passkey string) bool {
	return strings.Contains(announceURL, passkey)
}
