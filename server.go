package main

import (
	"fmt"
	"net"
	"sync"
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
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		//將message發送給全部的在線user
		this.mapLock.Lock()
		for _, client := range this.OnlineMap {
			client.C <- msg
		}
		this.mapLock.Unlock()
	}
}

//廣播訊息給全部的在線user
func (this *Server) BoardCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	//fmt.Println("連接成功.........")
	//用戶上線成功,加入OnlineMap中
	user := NewUser(conn)
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()
	//廣播當前用戶訊息
	this.BoardCast(user, " Online")
}

func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net listen error:", err)
	}

	//close  listen socket
	defer listener.Close()
	//啟動監聽Message的gorutine
	go this.ListenMessager()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept error:", err)
			continue
		}

		//do handler
		this.Handler(conn)
	}

}
