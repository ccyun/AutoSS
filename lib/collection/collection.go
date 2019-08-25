package collection

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

//Colle 采集
type Colle struct {
	conf RuleConf
	doc  *goquery.Document
}

//RuleConf 规则
type RuleConf struct {
	Page     string `json:"page"`
	ListRule string `json:"list_rule"`
	Attr     string `json:"attr"`
}

//NewCollection 新采集
func NewCollection(rule RuleConf) *Colle {
	c := new(Colle)
	c.conf = rule
	return c
}

//Page 采集页面
func (c *Colle) Page() error {
	res, err := http.Get(c.conf.Page)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}
	c.doc = doc
	return nil
}

//Attrs 采集列表
func (c *Colle) Attrs() []string {
	var data []string
	c.doc.Find(c.conf.ListRule).Each(func(i int, r *goquery.Selection) {
		if v, ok := r.Attr(c.conf.Attr); ok {
			data = append(data, v)
		}
	})
	return data
}
