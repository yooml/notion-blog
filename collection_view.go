package main

import (
	"encoding/json"
	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/caching_downloader"
)

type CollectionView struct {
	Id string
	Type string
	Name string
	Page_sort []string
}

func CollectionViewToPages(d *caching_downloader.Downloader) []string {
	s:=[]string{"cf7b1a3766ea499a90e568028152b10a"}
	//s1 := append(s, "cf7b1a3766ea499a90e568028152b10a")
	out,_:=d.Client.GetBlockRecords(s)
	r:=out.Results
	var pages []string
	pages=append(pages,"38e8c2a0cf7146a68cc24f847f94d800")
	for _,b:= range r {
		for c, e := range b.Block.ViewIDs {
			if c ==0||c==1{
				continue //跳过不想展示的前两个组
			}

			fliters := make([]*notionapi.QueryFilter, 1)
			var fliter *notionapi.QueryFilter = new(notionapi.QueryFilter)
			fliter.Comparator = "enum_contains"
			//fliter.ID = "f19ce6f4-1431-48e0-8390-766d7beab632"
			fliter.Type = "multi_select"
			for i := 0; i < 1; i++ {
				fliters[i] = fliter
				//fliters[i].ID = "f19ce6f4-1431-48e0-8390-766d7beab632"
				//fliters[i].Type = "multi_select"
			}
			var user *notionapi.User = new(notionapi.User)
			user.Locale = "zh-cn"
			user.TimeZone = "Asia/Shanghai"
			var query *notionapi.Query = new(notionapi.Query)
			query.Filter = fliters
			//query.CalendarBy="(WIG"
			query.FilterOperator = "and"
			coll, _ := d.Client.QueryCollection(b.Block.CollectionID, e, query, user)

			for _, g := range coll.RecordMap.CollectionViews {
				var bb CollectionView
				json.Unmarshal([]byte(g.Value), &bb)
				pages=append(pages,bb.Page_sort...)
			}

		}
	}
	//log.Println("pages", pages)
	return pages
}
