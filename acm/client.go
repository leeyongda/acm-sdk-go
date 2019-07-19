package acm

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/piex/transcode"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

/*
 * -----------------------------------
 * Life is short, you need Go
 * File: client.go
 * File Created: 2018-05-30 10:52:59 am
 * Author: coding
 * -----------------------------------
 * Last Modified: 2018-05-30  10:55:40 am
 * Modified By: coding
 * -----------------------------------
 */

var (
	// DefaultGroupName 默认分组
	DefaultGroupName = "DEFAULT_GROUP"
	DefaultNameSpace = ""
	Group            = "DEFAULT_GROUP"
	// 阿里云服务地址
	ServerUrl = "http://%s:8080/diamond-server/diamond"
)

var (
	Public string = "public"
	HZ     string = "hz"
	QD     string = "qd"
	SH     string = "sh"
	BJ     string = "bj"
	SZ     string = "sz"
)

// Params ...
type Params struct {
	Debug            bool   `json:"-"`
	AK               string `json:"-"`
	SK               string `json:"-"`
	ServerDomainAddr string `json:"-"` // 服务地址
	NameSpace        string `json:"dataId"`
	NameSpaceID      string `json:"tenant"`
	Group            string `json:"group,omitempty"`
	Type             string `json:"type,omitempty"`
	AppName          string `json:"appName,omitempty"`
	Desc             string `json:"desc,omitempty"`
	ConfigTags       string `json:"config_tags,omitempty"`
}

// 设置服务器地址
func (p *Params) SetServerDomainAddress(s string) {
	if p != nil {
		p.ServerDomainAddr = s
	}
}

// Acm ...
type Acm struct {
	V *viper.Viper
	m map[string]string
}

var once sync.Once
var m *Acm

// New ...
func New() *Acm {
	once.Do(func() {
		v := viper.New()
		v.SetConfigType("yaml")
		v.SetConfigFile("config.yaml")
		m = &Acm{V: v, m: make(map[string]string)}
		m.readConfig()
	})
	return m
}

func (a *Acm) readConfig() {
	if err := a.V.ReadInConfig(); err != nil {
		panic(err)
	}
	a.m = a.V.GetStringMapString("Server-Domain-Address")
}

// HmacSha HmacHsa 签名加密
func hmacSha(p Params, key string) string {

	hc := hmac.New(sha1.New, []byte(p.SK))
	hc.Write([]byte(key))
	sign := base64.StdEncoding.EncodeToString(hc.Sum([]byte(nil)))

	return sign
}

// 获取服务器 ip 地址
func (a *Acm) getServerIP(addr string) (string, error) {
	cl := &http.Client{
		Timeout: time.Second * 15,
	}
	addr = fmt.Sprintf(ServerUrl, addr)
	log.Println(addr)
	req, err := http.NewRequest("GET", addr, nil)
	res, err := cl.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "获取服务器ip失败")
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", errors.Wrap(err, "获取服务器ip失败")
	}

	return string(body), nil
}

func (a Acm) selectIP(p Params) string {
	var queryUrl string
	if p.Debug {
		queryUrl, _ = a.getServerIP(a.m["public"])
	} else {
		addr := a.m[strings.ToLower(p.ServerDomainAddr)]
		queryUrl, _ = a.getServerIP(addr)
	}
	return queryUrl
}

// Get ...
func (a *Acm) Get(p Params) ([]byte, error) {

	ok := CheckParams(p)
	if !ok {
		panic("参数不合法！")
	}

	dataID, group := a.ProcessCommonParams(p.NameSpace, p.Group)
	queryUrl := a.selectIP(p) + "/diamond-server/config.co?"
	timestamp := strconv.FormatInt(time.Now().Unix()*1000, 10)
	key := GroupKey(p.NameSpaceID, group, timestamp)
	sign := hmacSha(p, key)

	cl := &http.Client{}
	q := url.Values{}
	q.Set("dataId", dataID)
	q.Set("group", p.Group)
	q.Set("tenant", p.NameSpaceID)

	req, err := http.NewRequest("GET", queryUrl+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Spas-AccessKey", p.AK)
	req.Header.Set("timeStamp", timestamp)
	req.Header.Set("Spas-Signature", sign)
	res, err := cl.Do(req)

	if err != nil {
		return nil, errors.Wrap(err, "请求失败")

	}
	defer res.Body.Close()

	bodys, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return bodys, errors.Wrap(err, "读取失败")
	}
	// 转成gbk
	bodys = transcode.FromByteArray(bodys).Decode("GBK").ToByteArray()

	return bodys, nil

}

// PublishConfig .... 发布配置
func (a *Acm) PublishConfig(p Params, content string) ([]byte, string, error) {

	ok := CheckParams(p)
	if !ok {
		panic("参数不合法！")
	}
	if content == "" {
		panic("提交的配置内容不能为空!")
	}
	if p.NameSpaceID == "" {
		panic("命名空间ID不能为空!")
	}
	_, group := a.ProcessCommonParams(p.NameSpace, p.Group)
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	key := p.NameSpaceID + "+" + group + "+" + timestamp
	sign := hmacSha(p, key)
	queryurl := a.selectIP(p) + "/diamond-server/basestone.do?method=syncUpdateAll"
	content = transcode.FromString(content).Encode("GBK").ToString() // 转下编码格式
	result, err := json.Marshal(p)
	if err != nil {
		return nil, "", err
	}
	m := make(map[string]string, 0)
	errs := json.Unmarshal(result, &m)
	if errs != nil {
		return nil, "", errs
	}
	pp := url.Values{}
	for k, v := range m {
		pp.Set(k, v)
	}
	pp.Set("group", group)
	pp.Set("content", content)
	data := pp.Encode()
	cl := &http.Client{}
	req, err := http.NewRequest("POST", queryurl, strings.NewReader(data))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=GBK")
	req.Header.Set("Spas-AccessKey", p.AK)
	req.Header.Set("timeStamp", timestamp)
	req.Header.Set("Spas-Signature", sign)
	res, err := cl.Do(req)

	if err != nil {
		return nil, "", errors.Wrapf(err, "发布请求失败! StatusCode: %d Body: %s", res.StatusCode, "")
	}
	defer res.Body.Close()
	bodys, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, "", errors.Wrapf(err, "发布请求失败! StatusCode: %d Body: %s", res.StatusCode, "")
	}
	return bodys, res.Status, nil

}

// 获取命名空间的配置
func (a *Acm) GetAllConfig(p Params, pageNo, pageSize string) ([]byte, string, error) {

	ok := CheckParams(p)
	if !ok {
		panic("参数不合法！")
	}
	urls := a.selectIP(p) + "/diamond-server/basestone.do?method=getAllConfigByTenant&"
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	key := p.NameSpaceID + "+" + timestamp
	sign := hmacSha(p, key)
	pp := url.Values{}
	pp.Set("tenant", p.NameSpaceID)
	pp.Set("pageNo", pageNo)
	pp.Set("pageSize", pageSize)
	queryurl := pp.Encode()
	queryurl = urls + queryurl

	cl := &http.Client{}
	req, err := http.NewRequest("GET", queryurl, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Spas-AccessKey", p.AK)
	req.Header.Set("timeStamp", timestamp)
	req.Header.Set("Spas-Signature", sign)
	res, err := cl.Do(req)

	if err != nil {
		return nil, "", errors.Wrapf(err, "获取命名空间的配置失败! StatusCode: %d Body: %s", res.StatusCode, "")
	}
	defer res.Body.Close()
	bodys, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, "", errors.Wrapf(err, "获取命名空间的配置失败! StatusCode: %d Body: %s", res.StatusCode, "")
	}
	return bodys, res.Status, nil
}

// 删除配置
func (a *Acm) DeleteConfig(p Params) ([]byte, string, error) {

	ok := CheckParams(p)
	if !ok {
		panic("参数不合法！")
	}
	urls := a.selectIP(p) + "/diamond-server/datum.do?method=deleteAllDatums"
	dataID, group := a.ProcessCommonParams(p.NameSpace, p.Group)
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	key := p.NameSpaceID + "+" + group + "+" + timestamp
	sign := hmacSha(p, key)
	pp := url.Values{}
	pp.Set("dataId", dataID)
	pp.Set("group", group)
	pp.Set("tenant", p.NameSpaceID)
	cl := &http.Client{}
	req, err := http.NewRequest("POST", urls, strings.NewReader(pp.Encode()))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=GBK")
	req.Header.Set("Spas-AccessKey", p.AK)
	req.Header.Set("timeStamp", timestamp)
	req.Header.Set("Spas-Signature", sign)
	res, err := cl.Do(req)

	if err != nil {
		return nil, "", errors.Wrapf(err, "删除配置失败! StatusCode: %d Body: %s", res.StatusCode, "")
	}
	defer res.Body.Close()
	bodys, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, "", errors.Wrapf(err, "删除配置失败! StatusCode: %d Body: %s", res.StatusCode, "")
	}
	return bodys, res.Status, nil
}

// SaveConfigFile ...
func (c *Acm) SaveConfigFile(p Params, filetype string) {

	ok := CheckParams(p)
	if !ok {
		panic("参数不合法！")
	}
	content, err := c.Get(p)
	if err != nil {
		fmt.Println(err)
	}
	switch filetype {
	case "json":
		break
	case "yaml":
		break
	case "txt":
		break
	case "xml":
		break
	case "html":
		break
	case "properties":
		break

	default:
		panic("未知格式")
	}

	f, err := os.OpenFile(p.NameSpace+"."+filetype, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.Write(content)

}

// WatchConfig ... 监听配置文件
func (c *Acm) WatchConfig(p Params) {
}

// ProcessCommonParams 处理公共参数
func (c *Acm) ProcessCommonParams(dataID, group string) (string, string) {

	if group == "" {
		group = DefaultGroupName
	} else {
		group = strings.TrimSpace(group)
	}

	if dataID == "" || !IsValid(dataID) {
		panic("Invalid dataId.")
	}
	if !IsValid(group) {
		panic("Invalid group.")
	}

	return dataID, group

}
