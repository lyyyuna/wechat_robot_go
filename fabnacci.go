package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/mdp/qrterminal"
	"github.com/parnurzeal/gorequest"
)

type WeChat struct {
	tip          uint32
	deviceId     string
	req          *gorequest.SuperAgent
	uuid         string
	redirect_uri string
	base_uri     string
}

func init() {
	f, err := os.OpenFile("wechat_backend.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Error opening log file: %v", err)
	}

	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("WeChat Robot client begin.")
}

func NewWeChat() *WeChat {
	wechat := new(WeChat)
	wechat.tip = 0
	wechat.deviceId = "e000701000000000"
	wechat.req = gorequest.New().
		Set("User-Agent", "Mozilla/5.0 (X11; Linux i686; U;) Gecko/20070322 Kazehakase/0.4.5")
	return wechat
}

func (wechat *WeChat) getuuid() bool {
	log.Println("begin get uuid.")
	url := "https://login.weixin.qq.com/jslogin"

	_, body, errs := wechat.req.Get(url).
		Param("appid", "wx782c26e4c19acffb").
		Param("fun", "new").
		Param("lang", "zh_CN").
		Param("_", string(time.Now().Unix())).
		Param("redirect_uri", "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxnewloginpage").
		End()
	if errs != nil {
		log.Println(errs)
		return false
	}

	reg := regexp.MustCompile(`window.QRLogin.code = (\d+); window.QRLogin.uuid = "(\S+?)"`)
	matchedStr := reg.FindStringSubmatch(body)
	if matchedStr == nil {
		log.Println("No uuid can be found.")
		return false
	}
	code := matchedStr[1]
	wechat.uuid = matchedStr[2]
	if code == "200" {
		return true
	} else {
		return false
	}

}

func (wechat *WeChat) waitforlogin() string {
	url := "https://login.weixin.qq.com/cgi-bin/mmwebwx-bin/login"
	_, body, errs := wechat.req.Get(url).
		Param("tip", string(wechat.tip)).
		Param("uuid", wechat.uuid).
		Param("_", string(time.Now().Unix())).
		End()

	if errs != nil {
		return "500"
	}

	regx := regexp.MustCompile(`window.code=(\d+);`)
	matchedStr := regx.FindStringSubmatch(body)
	if matchedStr == nil {
		log.Println("No return code can be found.")
		return "500"
	}
	code := matchedStr[1]

	if code == "201" {
		log.Println("scan success")
		fmt.Println("成功扫描,请在手机上点击确认以登录")
	} else if code == "200" {
		log.Println("loginning")
		fmt.Println("正在登录")
		regx := regexp.MustCompile(`window.redirect_uri="(\S+?)";`)
		matchedStr := regx.FindStringSubmatch(body)
		if matchedStr == nil {
			log.Println("No redirect_uri can be found.")
			return "500"
		}
		wechat.redirect_uri = matchedStr[1]
	}

	return code

}

func main() {
	wechat := NewWeChat()
	success := wechat.getuuid()
	if success == false {
		log.Println("fail to get uuid.")
		fmt.Println("获取 uuid 失败")
		return
	}
	fmt.Println("请扫描二维码")
	qr_uuid := "https://login.weixin.qq.com/l/" + wechat.uuid
	qrterminal.Generate(qr_uuid, qrterminal.L, os.Stdout)

	for {
		wechat.waitforlogin()
		time.Sleep(time.Second * 1)
	}

}
