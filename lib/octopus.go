package lib

import (
	"AutoSS/lib/collection"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"os"
	"strconv"
	"strings"

	"github.com/tuotoo/qrcode"
)

//GuiConf GUI config
type GuiConf struct {
	guiConfigFile string
	ruleFile      string
	ruleConf      []collection.RuleConf
	data          map[string]interface{}
}

//Conf 配置账号
type Conf struct {
	Server     string `json:"server"`
	ServerPort uint64 `json:"server_port"`
	Password   string `json:"password"`
	Method     string `json:"method"`
	Plugin     string `json:"plugin"`
	PluginOpts string `json:"plugin_opts"`
	PluginArgs string `json:"plugin_args"`
	Remarks    string `json:"remarks"`
	Timeout    uint   `json:"timeout"`
}

//NewConfig 新配置
func NewConfig(guiConfigFile string, ruleFile string) (*GuiConf, error) {
	g := new(GuiConf)
	g.guiConfigFile = guiConfigFile
	if err := g.readConfig(&g.data); err != nil {
		return nil, err
	}
	g.data["configs"] = []interface{}{}

	g.ruleFile = ruleFile

	if err := g.readRule(&g.ruleConf); err != nil {
		return nil, err
	}

	return g, nil
}

//readConfig 读取配置
func (g *GuiConf) readConfig(data interface{}) error {
	var (
		err error
		s   []byte
	)
	if s, err = ioutil.ReadFile(g.guiConfigFile); err != nil {
		return err
	}
	if err = json.Unmarshal(s, data); err != nil {
		return err
	}
	return nil
}

//readRule 读取规则
func (g *GuiConf) readRule(data interface{}) error {
	var (
		err error
		s   []byte
	)
	if s, err = ioutil.ReadFile(g.ruleFile); err != nil {
		return err
	}
	if err = json.Unmarshal(s, data); err != nil {
		return err
	}
	return nil
}

//Save 保存配置
func (g *GuiConf) Save() {
	dstFile, _ := os.Create(g.guiConfigFile)
	str, _ := json.MarshalIndent(g.data, "", "    ")
	dstFile.WriteString(string(str))
	dstFile.Close()
}

//GetURLs 采集
func (g *GuiConf) GetURLs() error {
	configs := []*Conf{}
	for _, r := range g.ruleConf {
		c := collection.NewCollection(r)
		c.Page()
		for _, v := range c.Attrs() {
			if vv, err := g.Decode(fmt.Sprintf("%s%s", r.Page, v)); err == nil {
				configs = append(configs, vv)
			}
		}
	}
	g.data["configs"] = configs
	return nil
}

//Decode 解码
func (g *GuiConf) Decode(s string) (*Conf, error) {
	res, err := http.Get(s)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	m, err2 := qrcode.Decode(res.Body)
	if err2 != nil {
		return nil, err2
	}

	ss, err3 := base64.StdEncoding.DecodeString(strings.Replace(m.Content, "ss://", "", -1))
	if err3 != nil {
		return nil, err3
	}

	d := strings.FieldsFunc(string(ss), func(s rune) bool {
		switch s {
		case '\n', ':', '@':
			return true
		}
		return false
	})
	if len(d) == 4 {
		config := Conf{
			Server: d[2],
			// ServerPort: strconv.Atoi(d[3]),
			Password: d[1],
			Method:   d[0],
			Timeout:  5,
		}
		config.ServerPort, _ = strconv.ParseUint(d[3], 10, 64)
		return &config, nil
	}
	return nil, fmt.Errorf("no found")
}
