/*
 * Copyright 1999-2020 Alibaba Group Holding Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"bou.ke/monkey"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"net/http"
	"os"
	"os/signal"
)

var c = make(chan os.Signal, 1)

func init() {
	monkey.Patch(json.Marshal, func(v interface{}) ([]byte, error) {
		println("via monkey patch")
		return jsoniter.Marshal(v)
	})
}

func main() {

	sc := []constant.ServerConfig{
		{
			IpAddr: "192.168.1.222",
			Port:   8848,
		},
	}

	cc := constant.ClientConfig{
		NamespaceId:         "2b50d65a-f764-4472-b6c8-7eab6ed720c3", //namespace id
		TimeoutMs:           10000,
		NotLoadCacheAtStart: true,
		//LogDir:              "/tmp/nacos/log",
		//CacheDir:            "/tmp/nacos/cache",
		RotateTime: "1h",
		MaxAge:     3,
		LogLevel:   "info",
	}

	// 创建服务发现客户端的另一种方式 (推荐)
	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		panic(err)
	}

	success, err := namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          "192.168.1.33",
		Port:        8081,
		ServiceName: "service",
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"idc": "shanghai"},
		ClusterName: "DEFAULT",       // 默认值DEFAULT
		GroupName:   "DEFAULT_GROUP", // 默认值DEFAULT_GROUP
	})
	if err != nil {
		panic(err)
	}

	if success {
		fmt.Println("注册成功")
	}
	go func() {
		signal.Notify(c, os.Interrupt, os.Kill)
		s := <-c
		fmt.Println("Got signal:", s)
		bool, err := namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          "192.168.1.33",
			Port:        8081,
			ServiceName: "service",
			Ephemeral:   true,
			Cluster:     "DEFAULT",       // 默认值DEFAULT
			GroupName:   "DEFAULT_GROUP", // 默认值DEFAULT_GROUP
		})
		if err != nil {
			panic(err)
		}

		if bool {
			fmt.Println("注销成功")
		}
		os.Exit(0)
	}()
	defer func() {
		bool, err := namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          "192.168.1.33",
			Port:        8081,
			ServiceName: "service",
			Ephemeral:   true,
			Cluster:     "DEFAULT",       // 默认值DEFAULT
			GroupName:   "DEFAULT_GROUP", // 默认值DEFAULT_GROUP
		})
		if err != nil {
			panic(err)
		}

		if bool {
			fmt.Println("注销成功")
		}
	}()

	// 1.创建路由
	engine := gin.New()
	// 2.绑定路由规则，执行的函数
	// gin.Context，封装了request和response
	engine.GET("/", func(c *gin.Context) {

		c.JSON(http.StatusOK, gin.H{"status": "304"})
		return

	})
	// 3.监听端口，默认在8080
	// Run("里面不指定端口号默认为8080")
	engine.Run(":8081")

}
