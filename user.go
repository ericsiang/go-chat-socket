package main

import (
	"net"
	"strings"
)

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
	go user.ListenUserMessager()

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

////給當前用戶發送訊息
func (user *User) SendMessage(msg string) {
	user.conn.Write([]byte(msg))
}

//廣播發送訊息
func (user *User) DoMessage(msg string) {
	if msg == "who" {
		//查詢當前上線用戶
		user.server.mapLock.Lock()
		for _, tihsUser := range user.server.OnlineMap {
			//寫法一
			onlineMsg := "[" + tihsUser.Addr + "]" + tihsUser.Name + ":" + "Online Now...\n"
			user.SendMessage(onlineMsg)
		}
		user.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//訊息格式 => rename|newName
		newName := strings.Split(msg, "|")[1]
		//判斷用戶名是否存在
		_, ok := user.server.OnlineMap[newName]
		if ok {
			user.C <- "user name already used\n"
		} else {
			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()

			user.Name = newName
			user.SendMessage("update user name success\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//訊息格式 => to|userName|message

		spliteArr := strings.Split(msg, "|")
		toUserName := spliteArr[1]
		//1.獲取對方用戶名
		if toUserName == "" {
			user.SendMessage("format error ,the currect string is 'to|userName|message'\n")
			return
		}
		//2.依用戶名，獲取該user對象
		remoteUser, ok := user.server.OnlineMap[toUserName]
		if !ok {
			user.SendMessage("user not exist\n")
			return
		}
		//3.獲取預發送的訊息，通過該user對象將訊息發送過去
		toSendMsg := spliteArr[2]
		if toSendMsg == "" {
			user.SendMessage("message is empty\n")
			return
		}
		remoteUser.SendMessage(user.Name + " send message to you :" + toSendMsg)
	} else {
		//發送廣播訊息給當前上線用戶
		user.server.BoardCast(user, msg)
	}
}

//監聽當前user channel，一旦有消息，就立刻發送給客戶端
func (user *User) ListenUserMessager() {
	for {
		msg := <-user.C
		user.conn.Write([]byte(msg + "\n"))
	}
}
