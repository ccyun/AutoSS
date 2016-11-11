package octopus

import (
	"log"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

//Test_analysisJSON 测试json分析
func Test_analysisJSON(t *testing.T) {
	o := new(Octopus)
	doc, _ := goquery.NewDocument("https://api.mianvpn.com/ajax.php?verify=true&mod=getfreess")
	rule := Rule{
		URL:      "https://api.mianvpn.com/ajax.php?verify=true&mod=getfreess",
		Type:     "json",
		Server:   "i",
		Port:     "p",
		Method:   "m",
		Password: "pw",
		Auth:     false,
	}
	o.analysisJSON(rule, doc)
	log.Println(o.Config)
}

//Test_analysisJSON 测试json分析
func Test_analysisHTML(t *testing.T) {
	var err error
	log.Println("www.360sb.net---------------------------------------------------------------")
	o := new(Octopus)
	doc, _ := goquery.NewDocument("http://www.360sb.net")
	rule := Rule{
		URL:      "http://www.360sb.net/",
		Type:     "html",
		Filter:   ".*[:|：]|\\n|\\r|\\s",
		Block:    "section#fh5co-about>div.container div.col-md-4",
		Server:   "div.fh5co-person>span.fh5co-position:eq(1)",
		Port:     "span.fh5co-position:eq(2)",
		Method:   "span.fh5co-position:eq(3)",
		Password: "span.fh5co-position:eq(4)",
		Auth:     false,
	}
	o.analysisHTML(rule, doc)
	log.Println(o.Config)
	log.Println("www.getshadowsocks.com---------------------------------------------------------------")
	o = new(Octopus)
	doc, _ = goquery.NewDocument("http://www.getshadowsocks.com/")
	rule = Rule{
		URL:      "http://www.getshadowsocks.com/",
		Type:     "html",
		Filter:   ".*[:|：]|\\n|\\r|\\s",
		Block:    "#s",
		Server:   "p.lead:eq(0)",
		Port:     "p.lead:eq(1)",
		Method:   "p.lead:eq(3)",
		Password: "p.lead:eq(2)",
		Auth:     false,
	}
	o.analysisHTML(rule, doc)
	log.Println(o.Config)
	log.Println("www.ishadowsocks.org---------------------------------------------------------------")
	o = new(Octopus)
	doc, _ = goquery.NewDocument("http://www.ishadowsocks.org/")
	//log.Println(doc.Html())
	rule = Rule{
		URL:      "http://www.ishadowsocks.org/",
		Type:     "html",
		Filter:   ".*[:|：]|\\n|\\r|\\s",
		Block:    "div.container div.col-sm-4",
		Server:   "h4:eq(0)",
		Port:     "h4:eq(1)",
		Method:   "h4:eq(3)",
		Password: "h4:eq(2)",
		Auth:     false,
	}
	o.analysisHTML(rule, doc)
	log.Println(o.Config)
	log.Println("freeshadowsocks.cf---------------------------------------------------------------")
	o = new(Octopus)
	doc, err = goquery.NewDocument("http://freeshadowsocks.cf/")
	if err != nil {
		log.Println(err)
		return
	}
	//log.Println(doc.Html())
	rule = Rule{
		URL:      "http://freeshadowsocks.cf/",
		Type:     "html",
		Filter:   ".*[:|：]|\\n|\\r|\\s",
		Block:    "div.container div.col-md-6",
		Server:   "h4:eq(0)",
		Port:     "h4:eq(1)",
		Method:   "h4:eq(3)",
		Password: "h4:eq(2)",
		Auth:     false,
	}
	o.analysisHTML(rule, doc)
	log.Println(o.Config)
}

func Test_deqr(t *testing.T) {
	o := new(Octopus)
	o.deqr()
	log.Println(o.Config)
}
