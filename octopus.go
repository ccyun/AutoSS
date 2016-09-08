package octopus

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

//Octopus 定义爬虫结构体
type Octopus struct {
	Rule                                  []Rule
	Config                                Config
	RuleFile, ConfigFile, ShadowsocksFIle string
	remarks                               string
}

//Rule 定义规则结构体
type Rule struct {
	URL      string
	Type     string
	Filter   string
	Block    string
	Server   string
	Port     string
	Method   string
	Password string
	Auth     bool
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
func (o *Octopus) initRule() *Octopus {
	var (
		err  error
		data []byte
	)
	if data, err = ioutil.ReadFile(o.RuleFile); err != nil {
		log.Fatal(err)
	}
	if err = json.Unmarshal(data, &o.Rule); err != nil {
		log.Fatal(err)
	}
	return o
}

//run 并发执行
func (o *Octopus) Run() {
	o.initConfig()
	o.initRule()
	//启动协程并发运行
	runtime.GOMAXPROCS(runtime.NumCPU()) //读取cpu核数
	var wg sync.WaitGroup
	wg.Add(len(o.Rule))
	for _, v := range o.Rule {
		go o.analysis(&wg, v)
	}
	wg.Wait() //阻塞

	log.Println("成功获取到 " + strconv.Itoa(len(o.Config.Configs)) + " 个账号密码")
	o.saveConfig()
	o.reStart()
}

//analysis 分析DOM
func (o *Octopus) analysis(wg *sync.WaitGroup, rule Rule) {
	defer wg.Done()
	log.Println("正在获取：(" + rule.URL + ")")
	doc, err := goquery.NewDocument(rule.URL)
	if err != nil {
		log.Println(err.Error())
		return
	}
	switch rule.Type {
	case "html":
		if err := o.analysisHTML(rule, doc); err != nil {
			log.Println(err.Error())
			return
		}
	case "json":
		if err := o.analysisJSON(rule, doc); err != nil {
			log.Println(err.Error())
			return
		}
	}
}

//analysisHTML 分析html dom
func (o *Octopus) analysisHTML(rule Rule, doc *goquery.Document) error {
	var conf conf
	re, err := regexp.Compile(rule.Filter)
	if err != nil {
		return err
	}
	doc.Find(rule.Block).Each(func(i int, item *goquery.Selection) {
		conf.Server = re.ReplaceAllString(findExt(item, rule.Server), "")
		conf.ServerPort, _ = strconv.Atoi(re.ReplaceAllString(findExt(item, rule.Port), ""))
		conf.Password = re.ReplaceAllString(findExt(item, rule.Password), "")
		conf.Method = re.ReplaceAllString(findExt(item, rule.Method), "")
		if conf.Server != "" && conf.ServerPort != 0 && conf.Password != "" && conf.Method != "" {
			conf.Remarks = o.remarks
			conf.Auth = rule.Auth
			o.Config.Configs = append(o.Config.Configs, conf)
		}
	})
	return nil
}

//findExt 辅助规则
func findExt(item *goquery.Selection, rule string) string {
	re, _ := regexp.Compile(`(.*)\:(.*)\((.*)\)`)
	rules := re.FindStringSubmatch(rule)
	item = item.Find(rules[1])
	index, _ := strconv.Atoi(rules[3])
	switch rules[2] {
	case "eq":
		item = item.Eq(index)
	}
	return item.Text()
}

//analysisJSON 分析json
func (o *Octopus) analysisJSON(rule Rule, doc *goquery.Document) error {
	var (
		conf     conf
		jsonData []map[string]interface{}
	)
	if err := json.Unmarshal([]byte(doc.Text()), &jsonData); err != nil {
		return err
	}
	for _, v := range jsonData {
		conf.Server = v[rule.Server].(string)
		conf.ServerPort, _ = strconv.Atoi(v[rule.Port].(string))
		conf.Password = v[rule.Password].(string)
		conf.Method = v[rule.Method].(string)
		if conf.Server != "" && conf.ServerPort != 0 && conf.Password != "" && conf.Method != "" {
			conf.Remarks = o.remarks
			conf.Auth = rule.Auth
			o.Config.Configs = append(o.Config.Configs, conf)
		}
	}
	return nil
}

//saveConfig 保存配置文件
func (o *Octopus) saveConfig() {
	dstFile, _ := os.Create(o.ConfigFile)
	str, _ := json.MarshalIndent(o.Config, "", "    ")
	dstFile.WriteString(string(str))
	dstFile.Close()
}

//reStart 重启进程
func (o *Octopus) reStart() {
	log.Println("正在重新启动...")
	exec.Command("taskkill", "/im", "Shadowsocks.exe", "/f").Run()
	time.Sleep(time.Second * 2)
	exec.Command(o.ShadowsocksFIle).Start()
}
