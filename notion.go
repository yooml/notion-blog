package main

import (
	"fmt"
	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/caching_downloader"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"
)

var (
	nDownloadedPage = 0
	flgNoCache          bool
)
var (
	analyticsCode = "UA-194516-1"

	flgRedownloadNotion bool
	flgRedownloadPage   string
	flgDeploy           bool
	flgPreview          bool
	flgPreviewOnDemand  bool
	flgVerbose          bool
)



/*type Converter struct {
	page         *notionapi.Page
	notionClient *notionapi.Client
	galleries    [][]string

	r *tohtml.Converter
}

func (c *Converter) GenereateHTML() []byte {
	inner := string(c.r.ToHTML())
	page := c.page.Root()
	f := page.FormatPage()
	isMono := f != nil && f.PageFont == "mono"

	s := `<p></p>`
	if isMono {
		s += `<div style="font-family: monospace">`
	}
	s += inner
	if isMono {
		s += `</div>`
	}
	return []byte(s)
}*/

func WriteWithIoutil(name,content string) {
	data :=  []byte(content)
	if ioutil.WriteFile(name,data,0644) == nil {
		fmt.Println("写入文件成功:",content)
	}
}

func byteString(p []byte) string {
	for i := 0; i < len(p); i++ {
		if p[i] == 0 {
			return string(p[0:i])
		}
	}
	return string(p)
}

func runCaddy() {
	cmd := exec.Command("caddy", "-log", "stdout")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func preview() {
	go func() {
		time.Sleep(time.Second * 1)
		openBrowser("http://localhost:8080")
	}()
	runCaddy()
}

func rebuildAll(d *caching_downloader.Downloader) *Articles {
	regenMd()
	loadTemplates()
	articles := LoadArticles(d)
	readRedirects(articles)
	netlifyBuild(articles)
	return articles
}

func eventObserver(ev interface{}) {
	switch v := ev.(type) {
	case *caching_downloader.EventError:
		log.Printf(v.Error)
	case *caching_downloader.EventDidDownload:
		nDownloadedPage++
		log.Printf("%03d '%s' : downloaded in %s\n", nDownloadedPage, v.PageID, v.Duration)
	case *caching_downloader.EventDidReadFromCache:
		// TODO: only verbose
		nDownloadedPage++
		log.Printf("%03d '%s' : read from cache in %s\n", nDownloadedPage, v.PageID, v.Duration)
	case *caching_downloader.EventGotVersions:
		log.Printf("downloaded info about %d versions in %s\n", v.Count, v.Duration)
	}
}

func main()  {
	defer func() { // 必须要先声明defer，否则不能捕获到panic异常
		log.Println("d")
		if err := recover(); err != nil {
			log.Println(err) // 这里的err其实就是panic传入的内容
		}
		log.Println("e")
	}()
	os.MkdirAll("netlify_static", 0755)

	client := &notionapi.Client{}
	/*pageID := "19675f25c0fd4c41a704c87127861163"
	page, err := client.DownloadPage(pageID)
	subPages := page.GetSubPages()
	blocksToVisit := append([]string{}, page.Root().ContentIDs...)

	client.GetSubscriptionData("cf7b1a3766ea499a90e568028152b10a")
	log.Println("subPages:",subPages,blocksToVisit)
	if err != nil {
		log.Fatalf("DownloadPage() failed with %s\n", err)
	}
	//log.Println(page.NotionURL(),page.Root())
	r := tohtml.NewConverter(page)
	res := &Converter{
		notionClient: client,
		page:         page,
	}
	res.r = r
	log.Println(string(res.GenereateHTML()))*/
	//client := newNotionClient()
	cache, _ := caching_downloader.NewDirectoryCache(cacheDir)
	//must(err)
	d := caching_downloader.New(cache, client)
	/*s:=[]string{"cf7b1a3766ea499a90e568028152b10a"}
	//s1 := append(s, "cf7b1a3766ea499a90e568028152b10a")
	out,_:=d.Client.GetBlockRecords(s)
	r:=out.Results
	for a,b:= range r{
		log.Println("a,b:",a,b.Block.ID,b.Block.CollectionID)
		for c,e:=range b.Block.ViewIDs{
			//s:=[]string{d}
			//s1 := append(s, "cf7b1a3766ea499a90e568028152b10a")
			if c==2{
				fliters :=make([]*notionapi.QueryFilter,1)
				var fliter *notionapi.QueryFilter=new(notionapi.QueryFilter)
				fliter.Comparator = "enum_contains"
				//fliter.ID = "f19ce6f4-1431-48e0-8390-766d7beab632"
				fliter.Type = "multi_select"
				for i := 0; i < 1; i++ {
					fliters[i]=fliter
					//fliters[i].ID = "f19ce6f4-1431-48e0-8390-766d7beab632"
					//fliters[i].Type = "multi_select"
				}
				var user *notionapi.User=new(notionapi.User)
				user.Locale="zh-cn"
				user.TimeZone="Asia/Shanghai"
				var query *notionapi.Query=new(notionapi.Query) //=notionapi.Query{[],[],"(WIG","",fliters,[]}
				query.Filter=fliters
				//query.CalendarBy="(WIG"
				query.FilterOperator="and"
				coll,_:=d.Client.QueryCollection(b.Block.CollectionID,e,query,user)
				//outt,_:=d.Client.GetBlockRecords([]string{e})
				//rr:=outt.Results
				//toVisit := []string{coll.Result.BlockIDS}
				log.Println("rr[0].Block.ViewIDs:",coll.RawJSON,coll.RecordMap.CollectionViews)

				for h,g:=range coll.RecordMap.CollectionViews{
					var bb CollectionView
					dd:=json.Unmarshal([]byte(g.Value),&bb)
					log.Println("rr[0].Block.ViewIDs:",h,string(g.Value),dd,bb.Page_sort)
				}
			}
		}
	}*/
	//log.Println(out.RawJSON["results"])
	d.EventObserver = eventObserver
	d.RedownloadNewerVersions = true
	d.NoReadCache = flgNoCache

	rebuildAll(d)
}




