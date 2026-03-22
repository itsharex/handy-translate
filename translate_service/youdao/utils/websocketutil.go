package utils

import (
	"log/slog"
	"github.com/gorilla/websocket"
	neturl "net/url"
	"os"
	"strings"
	"sync"
)

/*
初始化websocket连接
*/
func InitConnection(url string) (*websocket.Conn, *sync.WaitGroup) {
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		slog.Error("WebSocket 连接失败", slog.Any("error", err))
		os.Exit(-1)
	}
	wg := sync.WaitGroup{}
	// 监听返回数据
	go messageHandler(ws, &wg)
	wg.Add(1)
	return ws, &wg
}

/*
初始化websocket连接, 并附带参数
*/
func InitConnectionWithParams(url string, paramsMap map[string][]string) (*websocket.Conn, *sync.WaitGroup) {
	params := neturl.Values{}
	for k, v := range paramsMap {
		params[k] = v
	}
	parseUrl, _ := neturl.Parse(url)
	parseUrl.RawQuery = params.Encode()
	return InitConnection(parseUrl.String())
}

/*
发送binary message
*/
func SendBinaryMessage(ws *websocket.Conn, message []byte) {
	ws.WriteMessage(websocket.BinaryMessage, message)
	slog.Debug("WebSocket 发送二进制消息", slog.Int("length", len(message)))
}

/*
发送text message
*/
func SendTextMessage(ws *websocket.Conn, message string) {
	ws.WriteMessage(websocket.TextMessage, []byte(message))
	slog.Debug("WebSocket 发送文本消息", slog.String("message", message))
}

func messageHandler(ws *websocket.Conn, wg *sync.WaitGroup) {
	for {
		msgType, msg, err := ws.ReadMessage()
		if err != nil {
			slog.Error("WebSocket 消息处理错误", slog.Any("error", err))
			break
		}
		switch msgType {
		case websocket.TextMessage:
			message := string(msg)
			slog.Debug("WebSocket 收到文本消息", slog.String("message", message))
			if !strings.Contains(message, "\"errorCode\":\"0\"") {
				wg.Done()
				os.Exit(-1)
			}
		case websocket.BinaryMessage:
			slog.Debug("WebSocket 收到二进制消息", slog.Int("length", len(msg)))
		case websocket.CloseMessage:
			slog.Info("WebSocket 连接已关闭", slog.String("message", string(msg)))
		}
	}
}
