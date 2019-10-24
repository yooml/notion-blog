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
	//readRedirects(articles)
	netlifyBuild(articles,d)
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
	/*defer func() { // 必须要先声明defer，否则不能捕获到panic异常
		log.Println("panic检测开始")
		if err := recover(); err != nil {
			log.Println(err) // 这里的err其实就是panic传入的内容
		}
		log.Println("panic检测结束")
	}()*/
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
	d.EventObserver = eventObserver
	d.RedownloadNewerVersions = true
	d.NoReadCache = flgNoCache

	rebuildAll(d)
}




