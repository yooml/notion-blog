package main

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/kjk/blog/tohtml"
	"github.com/kjk/notionapi"
)

// Converter renders article as html
type Converter struct {
	article      *Article
	page         *notionapi.Page
	notionClient *notionapi.Client
	idToArticle  func(string) *Article
	galleries    [][]string

	r *tohtml.Converter
}

func (c *Converter) maybeGetID(block *notionapi.Block) string {
	return notionapi.ToNoDashID(block.ID)
}

/*
func (c *Converter) WriteElement(block *notionapi.Block, tag string, attrs []string, content string, entering bool) {
	if !entering {
		if !isSelfClosing(tag) {
			c.r.Printf("</" + tag + ">")
			c.r.Newline()
		}
		return
	}

	s := "<" + tag
	nAttrs := len(attrs) / 2
	for i := 0; i < nAttrs; i++ {
		a := attrs[i*2]
		// TODO: quote value if necessary
		v := attrs[i*2+1]
		s += fmt.Sprintf(` %s="%s"`, a, v)
	}
	id := c.r.maybeGetID(block)
	if id != "" {
		s += ` id="` + id + `"`
	}
	s += ">"
	c.r.Printf(s)
	c.r.Newline()
	if len(content) > 0 {
		c.r.Printf(content)
		c.r.Newline()
	} else {
		c.r.RenderInlines(block.InlineContent)
	}
	c.r.Newline()
}
*/

// change https://www.notion.so/Advanced-web-spidering-with-Puppeteer-ea07db1b9bff415ab180b0525f3898f6
// =>
// /article/${id}
func (c *Converter) rewriteURL(uri string) string {
	id := notionapi.ExtractNoDashIDFromNotionURL(uri)
	if id == "" {
		return uri
	}
	article := c.idToArticle(id)
	// this might happen when I link to some-one else's public notion pages
	if article == nil {
		return uri
	}
	return article.URL()
}

func (c *Converter) getURLAndTitleForBlock(block *notionapi.Block) (string, string) {
	id := notionapi.ToNoDashID(block.ID)
	article := c.idToArticle(id)
	if article == nil {
		title := block.Title
		logf("No article for id %s %s\n", id, title)
		url := "/article/" + id + "/" + urlify(title)
		return url, title
	}

	return article.URL(), article.Title
}

func genGalleryMainHTML(galleryID int, imageURL string) string {
	s := `
  <div class="img-wrapper-wrapper">
    <div class="img-wrapper">
      <img id="id-gallery-{galleryID}" src="{imageURL}" />
      <a class="for-nav-icon nav-icon-left" href="#" onclick="imgPrev("{galleryID}"); return false;">
        <svg viewBox="0 0 24 24" preserveAspectRatio="xMidYMid meet" focusable="false" class="nav-icon">
          <g>
            <path d="M15.41 7.41L14 6l-6 6 6 6 1.41-1.41L10.83 12z" class="style-scope yt-icon">
            </path>
          </g>
        </svg>
      </a>

      <a class="for-nav-icon nav-icon-right" href="#" onclick="imgNext({galleryID}); return false;">
        <svg viewBox="0 0 24 24" preserveAspectRatio="xMidYMid meet" focusable="false" class="nav-icon" style="">
          <g>
            <path d="M10 6L8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z" class="yt-icon"></path>
          </g>
        </svg>
      </a>
    </div>
  </div>
`
	s = strings.Replace(s, "{galleryID}", strconv.Itoa(galleryID), -1)
	s = strings.Replace(s, "{imageURL}", imageURL, -1)
	return s
}

func genGalleryThumbHTML(galleryID int, n int, im *ImageMapping) string {
	s := `
    <div id="id-thumb-{galleryID}-{imageNo}" class="pa1 ib">
      <a href="#" onclick="changeShot({galleryID}, {imageNo}); return false;">
        <img id="id-thumb-img-{galleryID}-{imageNo}" src="{imageURL}" width="80" height="60" />
      </a>
	</div>
`
	s = strings.Replace(s, "{galleryID}", strconv.Itoa(galleryID), -1)
	ns := strconv.Itoa(n)
	s = strings.Replace(s, "{imageNo}", ns, -1)
	s = strings.Replace(s, "{imageURL}", im.relativeURL, -1)
	return s
}

func (c *Converter) renderGallery(block *notionapi.Block) bool {
	imageURLS := c.article.getGalleryImages(block)
	if len(imageURLS) == 0 {
		return false
	}
	panicIf(len(imageURLS) < 2, "expected gallery to have at least 2 images, got %d", len(imageURLS))
	galleryID := len(c.galleries)
	c.galleries = append(c.galleries, imageURLS)
	var images []*ImageMapping
	for _, link := range imageURLS {
		im := c.article.findImageMappingBySource(link)
		panicIf(im == nil, "didn't find ImageMapping for %s", link)
		images = append(images, im)
	}
	firstImage := images[0]
	s := genGalleryMainHTML(galleryID, firstImage.relativeURL)
	c.r.Printf(s)

	c.r.Printf(`<div class="center mt3 mb6">`)
	for i, im := range images {
		s := genGalleryThumbHTML(galleryID, i, im)
		c.r.Printf(s)
	}
	c.r.Printf(`</div>`)
	return true
}

// RenderImage renders BlockImage
func (c *Converter) RenderImage(block *notionapi.Block) bool {
	link := block.Source
	im := c.article.findImageMappingBySource(link)
	relURL := im.relativeURL
	imgURL := c.article.getImageBlockURL(block)
	if imgURL != "" {
		attrs := []string{"href", imgURL, "target", "_blank"}
		c.r.WriteElement(block, "a", attrs, "", true)
		{
			attrs2 := []string{"class", "blog-img", "src", relURL}
			c.r.WriteElement(block, "img", attrs2, "", true)
			c.r.WriteElement(block, "img", attrs2, "", false)
		}
		c.r.WriteElement(block, "a", attrs, "", false)
	} else {
		attrs := []string{"class", "blog-img", "src", relURL}
		c.r.WriteElement(block, "img", attrs, "", false)
		c.r.WriteElement(block, "img", attrs, "", true)
	}
	return true
}

// RenderPage renders BlockPage
func (c *Converter) RenderPage(block *notionapi.Block) bool {
	if c.r.Page.IsRoot(block) {
		c.r.Printf(`<div class="notion-page" id="%s">`, block.ID)
		c.r.RenderChildren(block)
		c.r.Printf(`</div>`)
		return true
	}

	cls := "page-link"
	if block.IsSubPage() {
		cls = "page"
	}

	url, title := c.getURLAndTitleForBlock(block)
	title = html.EscapeString(title)
	content := fmt.Sprintf(`<a href="%s">%s</a>`, url, title)
	attrs := []string{"class", cls}
	// title = html.EscapeString(title)
	c.r.WriteElement(block, "div", attrs, content, true)
	c.r.WriteElement(block, "div", attrs, content, false)
	return true
}

// RenderCode renders BlockCode
func (c *Converter) RenderCode(block *notionapi.Block) bool {
	// code := html.EscapeString(block.Code)
	// fmt.Fprintf(g.f, `<div class="%s">Lang for code: %s</div>
	// <pre class="%s">
	// %s
	// </pre>`, levelCls, block.CodeLanguage, levelCls, code)
	err := htmlHighlight(c.r.Buf, string(block.Code), block.CodeLanguage, "")
	panicIfErr(err)
	return true
}

// if returns false, the block will be rendered with default
func (c *Converter) blockRenderOverride(block *notionapi.Block) bool {
	if c.article.shouldSkipBlock(block) {
		return true
	}
	if c.renderGallery(block) {
		return true
	}
	switch block.Type {
	case notionapi.BlockPage:
		return c.RenderPage(block)
	case notionapi.BlockCode:
		return c.RenderCode(block)
	case notionapi.BlockImage:
		return c.RenderImage(block)
	}
	return false
}

// NewHTMLConverter returns new HTMLGenerator
func NewHTMLConverter(c *notionapi.Client, article *Article) *Converter {
	res := &Converter{
		notionClient: c,
		article:      article,
		page:         article.page,
	}

	r := tohtml.NewConverter(article.page)
	notionapi.PanicOnFailures = true
	r.RenderBlockOverride = res.blockRenderOverride
	r.RewriteURL = res.rewriteURL
	res.r = r

	return res
}

// Gen returns generated HTML
func (c *Converter) GenereateHTML() []byte {
	inner := string(c.r.ToHTML())
	page := c.page.Root()
	f := page.FormatPage()
	isMono := f != nil && f.PageFont == "mono"
	//isMono := page != nil

	s := `<p></p>`
	if isMono {
		s += `<div style="font-family: monospace">`
	}
	s += inner
	if isMono {
		s += `</div>`
	}
	return []byte(s)
}

func notionToHTML(client *notionapi.Client, article *Article, articles *Articles) ([]byte, []*ImageMapping) {
	//fmt.Printf("notionToHTML: %s\n", notionapi.ToNoDashID(article.ID))
	c := NewHTMLConverter(client, article)
	if articles != nil {
		c.idToArticle = func(id string) *Article {
			return articles.idToArticle[id]
		}
	}
	return c.GenereateHTML(), c.article.Images
}
