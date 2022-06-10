package main

import "net"

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

//創建一個用戶
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	//啟動監聽該user channel
	go user.ListenMessage()

	return user
}

//用戶上線
func (user *User) Online() {
	//用戶上線成功,加入OnlineMap中
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()
	//廣播當前用戶上線訊息
	user.server.BoardCast(user, " Online")
}

//用戶下線
func (user *User) Offline() {
	//用戶下線,從OnlineMap中移除
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()
	//廣播當前用戶上線訊息
	user.server.BoardCast(user, " Leave")
}

//廣播發送訊息
func (user *User) SendMessage(msg string) {
	//發送廣播訊息給當前上線用戶
	user.server.BoardCast(user, msg)
}

//監聽當前user channel，一旦有消息，就立刻發送給客戶端
func (user *User) ListenMessage() {
	for {
		msg := <-user.C
		user.conn.Write([]byte(msg + "\n"))
	}
}
