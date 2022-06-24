package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int64
	Name       string
	conn       net.Conn
	flag       int //當前client模式
}

func NewClient(serverIp string, serverPort int64) *Client {
	//建立客戶端對象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}
	//連接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net Dial Error", err.Error())
		return nil
	}

	client.conn = conn
	//返回對象
	return client

}

//處理server回應的消息，直接顯示到標準輸出中
func (client *Client) DelServerResponse() {
	//一旦client.conn有數據，就直接copy到os.Stdout標準輸出上，永久阻塞監聽
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write Error : ", err.Error())
		return
	}
}

func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.SelectUsers()
	fmt.Println(">>>>>>>>>請輸入聊天對象[用戶名]，exit退出>>>>>>>>>")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>>>>>>請輸入訊息，exit退出>>>>>>>>>")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			//訊息不為空時發送
			if len(chatMsg) > 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write Error : ", err.Error())
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>>>>>>>>請輸入訊息，exit退出>>>>>>>>>")
			fmt.Scanln(&chatMsg)
		}

		client.SelectUsers()
		fmt.Println(">>>>>>>>>請輸入聊天對象[用戶名]，exit退出>>>>>>>>>")
		fmt.Scanln(&remoteName)
	}
}

func (client *Client) PublicChat() {
	var chatMsg string
	fmt.Println(">>>>>>>>>請輸入訊息，exit退出>>>>>>>>>")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		//發給server

		//訊息不為空時發送
		if len(chatMsg) > 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write Error : ", err.Error())
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>>>>>>請輸入訊息，exit退出>>>>>>>>>")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println(">>>>>>>>>請輸入用戶名>>>>>>>>>")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write Error : ", err.Error())
		return false
	}
	return true
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {

		}

		//根據不同模式處理不同業務
		switch client.flag {
		case 1:
			client.PublicChat()
			break
		case 2:
			client.PrivateChat()
			break
		case 3:
			client.UpdateName()
			break
		}
	}
}

func (Client *Client) menu() bool {
	var flag int
	fmt.Println("0.退出")
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更改用戶名")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		Client.flag = flag
		return true
	} else {
		fmt.Println(">>>>>>>>>請輸入合法範圍內的數字>>>>>>>>>")
		return false
	}
}

var (
	serverIp   string
	serverPort int64
)

//./client.exe -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "設置服務器ip位址(默認是127.0.0.1)")
	flag.Int64Var(&serverPort, "port", 8888, "設置服務器port位(默認是8888)")
}

func main() {
	//命令行解析
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>>>>  連接服務器失敗")
		return
	}

	//處理server回應的goroutine
	go client.DelServerResponse()

	fmt.Println(">>>>>>>>  連接服務器成功")
	//啟動客戶端的業務
	client.Run()
}
