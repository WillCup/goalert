package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
	"bytes"
	"log"

	"encoding/base64"

	// ding "hello/ding"
	// ypclnt "github.com/yunpian/yunpian-go-sdk/sdk"
)

var YP YunPianInfo
var mailInfo MailInfo

const configFile = "./config.json"
var access_token string

type testItem struct {
	Name           string            `json:"name"`
	URL            string            `json:"url"`
	HttpProto      string            `json:"httpProto"`
	RequestTimeout string            `json:"requestTimeout"`
	ServerList     map[string]string `jsong:"serverList"`
	Interval       string            `json:"interval"`
}

type YunPianInfo struct {
	Enable    bool     `json:"enable"`
	APIKey    string   `json:"apiKey"`
	Phones    []string `json:"phones"`
	SmsPrefix string   `json:"smsPrefix"`
}

type MailInfo struct {
	Enable   bool     `json:"enable"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Smtp     string   `json:"smtp"`
	To       []string `json:"to"`
}

type config struct {
	AccessToken string 	 `json:"access_token"`
	MailInfo MailInfo    `json:"mail"`
	YPInfo   YunPianInfo `json:"YunPian"`
	TestItem []testItem  `json:"testItem"`
}

func main() {
	config := loadConfig(configFile)
	// for v, item := range config.TestItem {
	// 	log.Println(v)
	// 	log.Println(item)
	// }

	access_token = config.AccessToken
	// YP = config.YPInfo
	mailInfo = config.MailInfo
	for _, item := range config.TestItem {
		go itemTest(item)
	}

	//block main routin
	c := make(chan int)
	<-c

	// if ok, err := testServer("will_api.will.com.cn", "http", "/will_api_test.cgi?signKey=d28105da-4703-11e9-9a0f-f", "10"); err != nil {
	// 	log.Println("format")
	// 	log.Println(fmt.Errorf("接口访问出现问题", err))
	// } else {
	// 	log.Println(ok)
	// }
}

func itemTest(item testItem) {
	lastResult := map[string]bool{}
	lastFalseCount := map[string]int{}

	sig := make(chan int, 24)
	num := 0
	for {
		for ip, name := range item.ServerList {
			num++
			go func(ip, name string) {
				ok, err := testServer(ip, item.HttpProto, item.URL, item.RequestTimeout)
				if _, ok := lastResult[name]; !ok {
					lastResult[name] = true
				}
				if !ok {
					log.Printf("[%s]: %s\n", name, err)
					if lastResult[name] == true {
						SendMsg(fmt.Sprintf("[%s](%s) 请求接口失败:\n %s", name, ip, err))
						sendEmail(fmt.Sprintf("[%s](%s) 请求接口失败:\n %s", name, ip, err))
						lastFalseCount[name] += 1
						lastResult[name] = false
					} else {
						if v, ok := lastFalseCount[name]; ok {
							//存在
							if v > 2 {
								log.Printf("%s已经连续两次都没有访问成功了\n", name)
								SendMsg(fmt.Sprintf("[%s](%s) 请求接口失败:\n %s", name, ip, err))
								sendEmail(fmt.Sprintf("[%s](%s) 请求接口失败:\n %s", name, ip, err))
								lastFalseCount[name] = 1
							} else {
								log.Printf("%s还没达到次数2, 当前次数是%d，跳高\n", name, v)
								lastFalseCount[name] += 1
							}

						} else {
							log.Printf("%s首次触发告警\n", name)
							SendMsg(fmt.Sprintf("[%s](%s) 请求接口失败: \n %s", name, ip, err))
							sendEmail(fmt.Sprintf("[%s](%s) 请求接口失败:\n  %s", name, ip, err))
							lastFalseCount[name] = 1
						}
					}
				} else {
					log.Printf("[%s]: OK\n", name)
					if lastResult[name] == false {
						SendMsg(fmt.Sprintf("[%s](%s) 恢复正常", name, ip))
						sendEmail(fmt.Sprintf("[%s](%s) 恢复正常", name, ip))
						lastResult[name] = true
					}
				}
				sig <- 1
			}(ip, name)
		}
		for msg := range sig {
			if msg == 1 {
				num--
				if num <= 0 {
					break
				}
			}
		}
		t, _ := time.ParseDuration(item.Interval)
		time.Sleep(t)
	}
}

// 解析http返回的数据
func get_result(byt []byte, exp string) interface{} {
    var dat map[string]interface{}
    if len(string(byt)) == 0 {
		return "接口返回内容为空"			
	}
    if err := json.Unmarshal(byt, &dat); err != nil {
    	log.Println(string(byt))
        //panic(err)
        fmt.Sprintf("出问题了")
        return "返回结果出现问题，不是json串" + string(byt) 
    }
    // go语言不能像python的json.dump一样直接将json转为map嵌套结构，要一层一层弄。
    jsonObj, ok := dat["statusInfo"].(map[string]interface{})

    if !ok {
        fmt.Sprintf("出问题了")
        return dat["statusInfo"]
    } else {
        fmt.Sprintf("成功")
        return jsonObj["global"] //OK
    }  

}

type Ressult struct{
    status int
    statusInfo map[string]interface{}
}

func testServer(serverIP, proto, url, requestTimeout string) (bool, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	timeout, _ := time.ParseDuration(requestTimeout)
	client := &http.Client{Transport: tr, Timeout: timeout}
	resp, err := client.Get(fmt.Sprintf("%s://%s%s", proto, serverIP, url))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	result := get_result(body, "xxx")
	if resp.StatusCode >= 500 || result != "OK" {
		return false, errors.New(fmt.Sprintf("%s://%s%s \n status code(%d) \n 错误消息: %s \n", proto, serverIP, url, resp.StatusCode, result))
	} 

	// var f interface{}
	// err2 := json.Unmarshal(body, &f)
	// if err2 != nil {
	// 	return false, err
	// }
	// rspMap := f.(map[string]interface{})
	// if status, ok := rspMap["status"]; !ok || status.(string) == "fail" {
	// 	return false, errors.New(rspMap["error"].(string))
	// }

	// fmt.Print(string(body))
	return true, nil
}

func sendEmail(msg string) {
	log.Println("email sent !!! + " + msg)
	if !mailInfo.Enable {
		return
	}
	log.Printf("username :%s, smtp: %s" , mailInfo.Username, mailInfo.Smtp)
	// auth := smtp.PlainAuth("", mailInfo.Username, mailInfo.Password, mailInfo.Smtp)
	// to := mailInfo.To
	// nickname := "服务器小管家"
	// user := mailInfo.Username
	// subject := "服务器异常告警"
	// content_type := "Content-Type: text/plain; charset=UTF-8"
	// body := msg
	// context := []byte("To: " + strings.Join(to, ",") + "\r\nFrom: " + nickname +
	// 	"<" + user + ">\r\nSubject: " + subject + "\r\n" + content_type + "\r\n\r\n" + body)
	// err := smtp.SendMail(mailInfo.Smtp + ":25", auth, user, to, context)
	// if err != nil {
	// 	log.Printf("send mail error: %v", err)
	// }
	SendMail(mailInfo.Smtp +":25", mailInfo.Username, "服务器异常告警", msg, mailInfo.To, "服务器小管家")
	log.Printf("发送完毕 username :%s, smtp: %s" , mailInfo.Username, mailInfo.Smtp)
}

//ex: SendMail("127.0.0.1:25", (&mail.Address{"from name", "from@example.com"}).String(), "Email Subject", "message body", []string{(&mail.Address{"to name", "to@example.com"}).String()})
func SendMail(addr, from, subject, body string, to []string, nickname string) error {
	r := strings.NewReplacer("\r\n", "", "\r", "", "\n", "", "%0a", "", "%0d", "")

	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.Mail(r.Replace(from)); err != nil {
		return err
	}
	for i := range to {
		to[i] = r.Replace(to[i])
		if err = c.Rcpt(to[i]); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	msg := "To: " + strings.Join(to, ",") + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"Content-Transfer-Encoding: base64\r\n" +
		"\r\n" + base64.StdEncoding.EncodeToString([]byte(body))
	// msg := "To: " + strings.Join(to, ",") + "\r\nFrom: " + nickname +
	// 	"<" + user + ">\r\nSubject: " + subject + "\r\n" + 
	// 	"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
	// 	"Content-Transfer-Encoding: base64\r\n" +
	// 	"\r\n" + "\r\n\r\n" + EncodeToString([]byte(body))

	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}


func SendMsg(msg string) {
	log.Println("SendMsg world!!! + " + msg)
	url := fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s", access_token)
    log.Println("URL:>", url)
    mm := fmt.Sprintf(`{"msgtype":"text","text":{"content":"%s"},"at":{"isAtAll":"true"}}`, msg)
    var jsonStr = []byte(mm)
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
    req.Header.Set("X-Custom-Header", "myvalue")
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    log.Println("response Status:", resp.Status)
    log.Println("response Headers:", resp.Header)
    body, _ := ioutil.ReadAll(resp.Body)
    log.Println("response Body:", string(body))
	
}


func loadConfig(fileName string) *config {
	file, _ := os.Open(fileName)
	defer file.Close()
	decoder := json.NewDecoder(file)
	conf := new(config)
	err := decoder.Decode(&conf)
	if err != nil {
		log.Println("Error:", err)
	}
	return conf
}