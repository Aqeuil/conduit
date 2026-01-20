package protocols

import (
	"fmt"
	"net"

	"github.com/spf13/viper"
	"go.conduit.cn/conduit/v2/server/config"
)

// StartListener 监听端口
func StartListener() {
	// 1. 监听tcp
	listener, err := net.Listen(viper.Get(config.ServerNetwork).(string), ":"+viper.Get(config.ServerPort).(string))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err2 := listener.Accept()
		if err2 != nil {
			continue
		}
		go handleConnection(conn)
	}

	// 2. 解析包通过protocols判断协议类型
	// 3. 解析http协议
	// 4. 路由去core完成应用匹配
	// 5. 根据应用去app-config进行拦截器处理
	// 6. core转发
	// 7. 根据应用去app-config进行后置拦截器处理
	// 8. 返回
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// todo
	// 解析协议头判断协议类型
	// 处理http 帧?包?
	// 解析内容 (头, 路由, **)
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Failed to read request:", err)
		return
	}

	request := string(buffer[:n])
	fmt.Println("Received request:", request)

	response := "HTTP/1.1 200 OK\r\nContent-Length: 12\r\n\r\nHello, World!"
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Failed to send response:", err)
		return
	}

	fmt.Println("Sent response:", response)
}
