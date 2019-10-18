package main

import (
	"html/template"
	"log"
	"sort"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/caching_downloader"
)

var (
	notionBlogsStartPage      = "8ba29f51a5af4fe3a02bc4f5286d626a"
	notionWebsiteStartPage    = "19675f25c0fd4c41a704c87127861163"
	notionGoCookbookStartPage = "cf7b1a3766ea499a90e568028152b10a"
)
/*
var (
	notionBlogsStartPage      = "300db9dc27c84958a08b8d0c37f4cfe5"
	notionWebsiteStartPage    = "568ac4c064c34ef6a6ad0b8d77230681"
	notionGoCookbookStartPage = "7495260a1daa46118858ad2e049e77e6"
)
*/

// Articles has info about all articles downloaded from notion
type Articles struct {
	idToArticle map[string]*Article
	idToPage    map[string]*notionapi.Page
	// all downloaded articles
	articles []*Article
	// articles that are not hidden
	articlesNotHidden []*Article
	// articles that belong to a blog
	blog []*Article
	// blog articles that are not hidden
	blogNotHidden []*Article
}

func (a *Articles) getNotHidden() []*Article {
	if a.articlesNotHidden == nil {
		var arr []*Article
		for _, article := range a.articles {
			if !article.IsHidden() {
				arr = append(arr, article)
			}
		}
		a.articlesNotHidden = arr
	}
	return a.articlesNotHidden
}

func (a *Articles) getBlogNotHidden() []*Article {
	if a.blogNotHidden == nil {
		var arr []*Article
		for _, article := range a.blog {
			if !article.IsHidden() {
				arr = append(arr, article)
			}
		}
		a.blogNotHidden = arr
	}
	return a.blogNotHidden
}

func buildArticleNavigation(article *Article, isRootPage func(string) bool, idToBlock map[string]*notionapi.Block) {
	// some already have path (e.g. those that belong to a collection)
	if len(article.Paths) > 0 {
		return
	}

	page := article.page.Root()
	currID := normalizeID(page.ParentID)

	var paths []URLPath
	for !isRootPage(currID) {
		block := idToBlock[currID]
		if block == nil {
			break
		}
		// parent could be a column
		if block.Type != notionapi.BlockPage {
			currID = normalizeID(block.ParentID)
			continue
		}
		title := block.Title
		uri := "/article/" + normalizeID(block.ID) + "/" + urlify(title)
		path := URLPath{
			Name: title,
			URL:  uri,
		}
		paths = append(paths, path)
		currID = normalizeID(block.ParentID)
	}

	// set in reverse order
	n := len(paths)
	for i := 1; i <= n; i++ {
		path := paths[n-i]
		article.Paths = append(article.Paths, path)
	}
}

func normalizeID(id string) string {
	return notionapi.ToNoDashID(id)
}

func addIDToBlock(block *notionapi.Block, idToBlock map[string]*notionapi.Block) {
	id := normalizeID(block.ID)
	idToBlock[id] = block
	for _, block := range block.Content {
		if block == nil {
			continue
		}
		addIDToBlock(block, idToBlock)
	}
}

// build navigation bread-crumbs for articles
func buildArticlesNavigation(articles *Articles) {
	idToBlock := map[string]*notionapi.Block{}
	for _, a := range articles.articles {
		page := a.page
		if page == nil {
			continue
		}
		addIDToBlock(page.Root(), idToBlock)
	}

	isRoot := func(id string) bool {
		id = notionapi.ToNoDashID(id)
		switch id {
		case notionBlogsStartPage, notionWebsiteStartPage, notionGoCookbookStartPage:
			return true
		}
		return false
	}

	for _, article := range articles.articles {
		buildArticleNavigation(article, isRoot, idToBlock)
	}
}

func LoadArticles(d *caching_downloader.Downloader) *Articles {
	res := &Articles{}
	//_, err := d.DownloadPagesRecursively(notionWebsiteStartPage)
	_, err := downloadPagesRecursively(d,notionWebsiteStartPage)
	must(err)
	res.idToPage = d.IdToPage
	//log.Println("res.idToPage:",res.idToPage["ee6bb471a110437ebe54dcdc483ba978"])

	c := d.GetClientCopy()
	res.idToArticle = map[string]*Article{}
	for id, page := range res.idToPage {
		log.Println(id)
		panicIf(id != notionapi.ToNoDashID(id), "bad id '%s' sneaked in", id)
		article := notionPageToArticle(c, page)
		if article.urlOverride != "" {
			verbose("url override: %s => %s\n", article.urlOverride, article.ID)
		}
		res.idToArticle[id] = article
		// this might be legacy, short id. If not, we just set the same value twice
		articleID := article.ID
		res.idToArticle[articleID] = article
		if article.IsBlog() {
			res.blog = append(res.blog, article)
		}
		res.articles = append(res.articles, article)
	}

	for _, article := range res.articles {
		html, images := notionToHTML(c, article, res)
		article.BodyHTML = string(html)
		article.HTMLBody = template.HTML(article.BodyHTML)
		article.Images = append(article.Images, images...)
	}

	buildArticlesNavigation(res)

	sort.Slice(res.blog, func(i, j int) bool {
		return res.blog[i].PublishedOn.After(res.blog[j].PublishedOn)
	})
	log.Println(res.articles[0].ID)


	return res
}

// MonthArticle combines article and a month
type MonthArticle struct {
	*Article
	DisplayMonth string
}

// Year describes articles in a given year
type Year struct {
	Name     string
	Articles []MonthArticle
}

// DisplayTitle returns a title for an article
func (a *MonthArticle) DisplayTitle() string {
	if a.Title != "" {
		return a.Title
	}
	return "no title"
}

// NewYear creates a new Year
func NewYear(name string) *Year {
	return &Year{Name: name, Articles: make([]MonthArticle, 0)}
}

func buildYearsFromArticles(articles []*Article) []Year {
	res := make([]Year, 0)
	var currYear *Year
	var currMonthName string
	n := len(articles)
	for i := 0; i < n; i++ {
		a := articles[i]
		yearName := a.PublishedOn.Format("2006")
		if currYear == nil || currYear.Name != yearName {
			if currYear != nil {
				res = append(res, *currYear)
			}
			currYear = NewYear(yearName)
			currMonthName = ""
		}
		ma := MonthArticle{Article: a}
		monthName := a.PublishedOn.Format("01")
		if monthName != currMonthName {
			ma.DisplayMonth = a.PublishedOn.Format("January 2")
		} else {
			ma.DisplayMonth = a.PublishedOn.Format("2")
		}
		currMonthName = monthName
		currYear.Articles = append(currYear.Articles, ma)
	}
	if currYear != nil {
		res = append(res, *currYear)
	}
	return res
}

func filterArticlesByTag(articles []*Article, tag string, include bool) []*Article {
	res := make([]*Article, 0)
	for _, a := range articles {
		hasTag := false
		for _, t := range a.Tags {
			if tag == t {
				hasTag = true
				break
			}
		}
		if include && hasTag {
			res = append(res, a)
		} else if !include && !hasTag {
			res = append(res, a)
		}
	}
	return res
}

func downloadPagesRecursively(d *caching_downloader.Downloader,pageID string) ([]*notionapi.Page, error) {
	toVisit := []string{pageID}
	downloaded := map[string]*notionapi.Page{}
	for len(toVisit) > 0 {
		pageID := notionapi.ToNoDashID(toVisit[0])
		toVisit = toVisit[1:]
		if downloaded[pageID] != nil {
			continue
		}

		page, err := d.DownloadPage(pageID)
		if err != nil {
			return nil, err
		}
		downloaded[pageID] = page

		subPages := getSubPages(page)
		toVisit = append(toVisit, subPages...)
	}
	n := len(downloaded)
	if n == 0 {
		return nil, nil
	}
	var ids []string
	for id := range downloaded {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	pages := make([]*notionapi.Page, n)
	for i, id := range ids {
		pages[i] = downloaded[id]
	}
	return pages, nil
}

// GetSubPages return list of ids for direct sub-pages of this page
func getSubPages(p *notionapi.Page) []string {
	root := p.Root()
	//panicIf(!isPageBlock(root))
	subPages := map[string]struct{}{}
	seenBlocks := map[string]struct{}{}
	blocksToVisit := append([]string{}, root.ContentIDs...)
	for len(blocksToVisit) > 0 {
		id := notionapi.ToDashID(blocksToVisit[0])
		blocksToVisit = blocksToVisit[1:]
		if _, ok := seenBlocks[id]; ok {
			continue
		}
		seenBlocks[id] = struct{}{}
		block := p.BlockByID(id)
		subPages[id] = struct{}{}
		//if p.IsSubPage(block) {
		//	subPages[id] = struct{}{}
		//}
		// need to recursively scan blocks with children
		blocksToVisit = append(blocksToVisit, block.ContentIDs...)
	}
	res := []string{}
	for id := range subPages {
		res = append(res, id)
	}
	sort.Strings(res)
	return res
}