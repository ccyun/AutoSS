package octopus

import (
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

//Octopus 定义爬虫结构体
type Octopus struct {
	Files   []string
	Config  Config
	remarks string
}

//Config 定义导出配置结构体
type Config struct {
	Configs      []conf      `json:"configs"`
	Strategy     string      `json:"strategy"`
	Index        int         `json:"index"`
	Global       bool        `json:"global"`
	Enabled      bool        `json:"enabled"`
	ShareOverLan bool        `json:"shareOverLan"`
	IsDefault    bool        `json:"isDefault"`
	LocalPort    int         `json:"localPort"`
	PacURL       interface{} `json:"pacUrl"`
	UseOnlinePac bool        `json:"useOnlinePac"`
}
type deqr struct {
	Status int `json:"status"`
	Data   struct {
		RawData string
	} `json:"data"`
}

//conf 定义配置项
type conf struct {
	Server     string `json:"server"`
	ServerPort int    `json:"server_port"`
	Password   string `json:"password"`
	Method     string `json:"method"`
	Remarks    string `json:"remarks"`
	Auth       bool   `json:"auth"`
}

//initConfig 初始化配置
func (o *Octopus) initConfig() *Octopus {
	o.Config.Strategy = "com.shadowsocks.strategy.ha"
	o.Config.Index = -1
	o.Config.Enabled = true
	o.Config.LocalPort = 1080
	o.remarks = time.Unix(time.Now().Unix(), 0).Format("01/02 15:04")
	return o
}

//initRule 初始化规则
func (o *Octopus) initRule() error {
	var (
		err  error
		data []byte
	)
	if data, err = ioutil.ReadFile("rule.json"); err != nil {
		log.Fatal(err)
	}
	if err = json.Unmarshal(data, &o.Files); err != nil {
		log.Fatal(err)
	}
	return nil
}

//Run 并发执行
func (o *Octopus) Run() {
	o.initConfig()
	o.initRule()
	//启动协程并发运行
	runtime.GOMAXPROCS(runtime.NumCPU()) //读取cpu核数
	var wg sync.WaitGroup
	wg.Add(len(o.Files))
	for _, v := range o.Files {
		go func(v string) {
			o.deqr(v)
			wg.Done()
		}(v)
	}
	wg.Wait() //阻塞
	log.Println("成功获取到 " + strconv.Itoa(len(o.Config.Configs)) + " 个账号密码")
	o.saveConfig()
	o.reStart()
}

//Request curl请求
func Request(method string, url string, body io.Reader, contentType string) (int, []byte, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return -1, nil, fmt.Errorf("construct http request failed, requrl = %s, err:%s", url, err.Error())
	}
	if method == "POST" {
		switch contentType {
		case "json":
			request.Header.Add("Content-Type", "application/json")
		case "form":
			request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		}
	}
	client := &http.Client{}
	response, err := client.Do(request)
	if response != nil {
		defer response.Body.Close()
		var respBody []byte
		switch response.Header.Get("Content-Encoding") {
		case "gzip":
			reader, _ := gzip.NewReader(response.Body)
			defer reader.Close()
			respBody, _ = ioutil.ReadAll(reader)
		default:
			respBody, _ = ioutil.ReadAll(response.Body)
		}
		if response.StatusCode > 300 {
			return response.StatusCode, respBody, fmt.Errorf("http request fail, url: %s", url)
		}
		return response.StatusCode, respBody, nil
	}
	if err != nil {
		return -1, nil, fmt.Errorf("http request fail, url: %s, error:%s", url, err.Error())
	}
	return -1, nil, fmt.Errorf("http request fail, url: %s, error:%s", url, err.Error())
}

func (o *Octopus) deqr(file string) {
	_, res, err := Request("GET", "http://cli.im/Api/Browser/deqr?data="+file, strings.NewReader(""), "")
	if err != nil {
		return
	}
	var info deqr
	if err = json.Unmarshal(res, &info); err != nil {
		return
	}

	info.Data.RawData = strings.Replace(info.Data.RawData, "ss://", "", 10)

	dataInfo, _ := base64.StdEncoding.DecodeString(info.Data.RawData)
	if info.Data.RawData != "" {
		info.Data.RawData = string(dataInfo)
		log.Println(info.Data.RawData)
		d := strings.FieldsFunc(info.Data.RawData, func(s rune) bool {
			switch s {
			case '\n', 0x3A, 0x40:
				return true
			}
			return false
		})
		if d[0] != "" && d[1] != "" && d[2] != "" && d[3] != "" {
			var conf conf
			conf.Server = d[2]
			conf.ServerPort, _ = strconv.Atoi(d[3])
			conf.Password = d[1]
			conf.Method = d[0]
			conf.Remarks = o.remarks
			conf.Auth = false
			o.Config.Configs = append(o.Config.Configs, conf)
		}
	}
}

//saveConfig 保存配置文件
func (o *Octopus) saveConfig() {
	dstFile, _ := os.Create("gui-config.json")
	str, _ := json.MarshalIndent(o.Config, "", "    ")
	dstFile.WriteString(string(str))
	dstFile.Close()
}

//reStart 重启进程
func (o *Octopus) reStart() {
	log.Println("正在重新启动...")
	exec.Command("taskkill", "/im", "Shadowsocks.exe", "/f").Run()
	time.Sleep(time.Second * 2)
	exec.Command("Shadowsocks.exe").Start()
}
