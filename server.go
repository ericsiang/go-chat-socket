package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int64
	//在線用戶的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	//訊息廣播的channel
	Message chan string
}

//創建server
func NewSever(ip string, port int64) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

//監聽Message廣播的channel的goroutine，一旦有訊息需發送給OnlineMap全部的在線user
func (server *Server) ListenBoardCastMessager() {
	for {
		msg := <-server.Message
		//將message發送給全部的在線user
		server.mapLock.Lock()
		for _, client := range server.OnlineMap {
			client.C <- msg
		}
		server.mapLock.Unlock()
	}
}

//廣播訊息給全部的在線user
func (server *Server) BoardCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	server.Message <- sendMsg
}

func (server *Server) Handler(conn net.Conn) {
	//fmt.Println("連接成功.........")
	user := NewUser(conn, server)
	user.Online()
	//用戶是否活躍的channel
	isLive := make(chan bool)

	//接受客戶端發送的訊息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			//表示客戶端close
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read Error:", err)
				return
			}
			//提取用戶的消息(去除'\n')
			msg := string(buf[:n-1])
			//將取到的訊息進行訊息處理
			user.DoMessage(msg)
			//用戶有發送訊息，表示活耀
			isLive <- true
		}
	}()

	for {
		select {
		case <-isLive:
		case <-time.After(time.Second * 10):
			//已超時,
			//將當前用戶強制關閉
			user.SendMessage("Your are been kick")

			//關閉資源
			close(user.C)
			//關閉連接
			conn.Close()
			//退出當前Handler
			return
		}
	}
}

func (server *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("net listen error:", err)
	}

	//close  listen socket
	defer listener.Close()
	//啟動監聽Message的gorutine
	go server.ListenBoardCastMessager()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept error:", err)
			continue
		}

		//do handler
		server.Handler(conn)
	}

}
