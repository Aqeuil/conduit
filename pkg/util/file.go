package util

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// CreatePathDir 判断文件路径的文件夹是否存在，不存在则创建文件夹
func CreatePathDir(path string) (bool, error) {
	dir := filepath.Dir(path)
	_, err := os.Stat(dir)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return false, err
		} else {
			return true, nil
		}
	}

	return false, err
}

// Download 下载文件到指定目录
func Download(dir, url string) (string, error) {
	// 确保目录存在
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 发起 HTTP 请求
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("服务器返回错误状态: %s", resp.Status)
	}

	// 从 URL 或响应头提取文件名
	filename, err := ExtractFilename(url, resp.Header.Get("Content-Disposition"))
	if err != nil {
		return "", fmt.Errorf("无法确定文件名")
	}

	// 构建完整文件路径
	filePath := filepath.Join(dir, filename)

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 写入数据
	if _, err := io.Copy(file, resp.Body); err != nil {
		// 如果写入失败，删除已创建的文件
		_ = os.Remove(filePath)
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	return filePath, nil
}

// ExtractFilename 从 URL 和 Content-Disposition 头提取文件名
func ExtractFilename(urlStr, contentDisposition string) (string, error) {
	// 优先从 Content-Disposition 头提取
	if contentDisposition != "" {
		if strings.HasPrefix(contentDisposition, "attachment; filename=") {
			filename := strings.TrimPrefix(contentDisposition, "attachment; filename=")
			return strings.Trim(filename, "\""), nil
		}
	}

	// 从 URL 路径提取
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	// 获取URL的路径部分
	urlPath := parsedURL.Path

	// 如果路径以斜杠结尾，则去除末尾斜杠
	urlPath = strings.TrimSuffix(urlPath, "/")

	// 使用path包提取基础文件名
	filename := path.Base(urlPath)

	// 如果基础文件名是根路径（如"/"），则返回空字符串
	if filename == "/" || filename == "." {
		return "", nil
	}

	return filename, nil
}

// DownloadImageToBase64 将图片地址转位base64
func DownloadImageToBase64(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image, status code: %d", resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %v", err)
	}

	base64Str := base64.StdEncoding.EncodeToString(imageData)
	return base64Str, nil
}
