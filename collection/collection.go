package collection

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tuotoo/qrcode"
)

//GuiConf GUI config
type GuiConf struct {
	guiConfigFile string
	ruleFile      string
	ruleConf      []RuleConf
	data          map[string]interface{}
}

//RuleConf 规则
type RuleConf struct {
	URL  string `json:"url"`
	List string `json:"list"`
	Attr string `json:"attr"`
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

//HTTPCurl Get 请求
func HTTPCurl(url string) (io.Reader, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	s, _ := ioutil.ReadAll(res.Body)
	return bytes.NewReader(s), nil
}

//ReadJSONConfig 读取配置
func ReadJSONConfig(file string, data interface{}) error {
	var (
		err error
		s   []byte
	)
	if s, err = ioutil.ReadFile(file); err != nil {
		return err
	}
	if err = json.Unmarshal(s, data); err != nil {
		return err
	}
	return nil
}

//NewConfig 新配置
func NewConfig(guiConfigFile string, ruleFile string) (*GuiConf, error) {
	g := new(GuiConf)
	g.guiConfigFile = guiConfigFile
	if err := ReadJSONConfig(guiConfigFile, &g.data); err != nil {
		return nil, err
	}
	g.data["configs"] = []interface{}{}
	g.ruleFile = ruleFile
	if err := ReadJSONConfig(ruleFile, &g.ruleConf); err != nil {
		return nil, err
	}
	return g, nil
}

//Save 保存配置
func (g *GuiConf) Save() {

	dstFile, _ := os.Create(g.guiConfigFile)
	str, _ := json.MarshalIndent(g.data, "", "    ")
	dstFile.WriteString(string(str))

	dstFile.Close()
}

//GetURLs 采集
func (g *GuiConf) GetURLs() (int, error) {
	configs := []*Conf{}
	for _, r := range g.ruleConf {
		log.Println("采集网页：", r.URL)
		res, err := HTTPCurl(r.URL)
		if err != nil {
			continue
		}
		doc, err2 := goquery.NewDocumentFromReader(res)
		if err2 != nil {
			log.Println(err2)
			continue
		}

		doc.Find(r.List).Each(func(i int, s *goquery.Selection) {
			if v, ok := s.Attr(r.Attr); ok {
				if vv, err := g.QRDecode(fmt.Sprintf("%s%s", r.URL, v)); err == nil {
					configs = append(configs, vv)
				}
			}
		})
	}
	g.data["configs"] = configs
	return len(configs), nil
}

//QRDecode 解码
func (g *GuiConf) QRDecode(s string) (*Conf, error) {
	res, err := HTTPCurl(s)
	if err != nil {
		return nil, err
	}
	m, err2 := qrcode.Decode(res)
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
