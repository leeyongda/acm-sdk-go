package main

import (
	"bytes"
	"fmt"
	"github.com/leeyongda/acm-sdk-go/acm"
	"log"
	"os"

	"github.com/spf13/viper"
)

var v *viper.Viper

func main() {

	p := acm.Params{
		Debug:            true, // 走公网
		ServerDomainAddr: acm.Public,
		AK:               os.Getenv("AK"),
		SK:               os.Getenv("SK"),
		NameSpace:        "dev.test.com.app",
		NameSpaceID:      "xxx",
		Group:            "",
		Type:             "yaml",
		AppName:          "app",
		Desc:             "app",
		ConfigTags:       "app",
	}
	a := acm.New()
	// 获取配置
	result, err := a.Get(p)
	if err != nil {
		fmt.Println(err)
		return
	}

	v = viper.New()
	v.SetConfigType("yaml")
	fmt.Println(string(result))
	if err := v.ReadConfig(bytes.NewBuffer(result)); err != nil {
		log.Println(err)
		return
	}
	fmt.Println(v.GetString("key"))

}
