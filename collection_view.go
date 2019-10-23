package main

import (
	"encoding/json"
	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/caching_downloader"
	"log"
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
	for a,b:= range r {
		log.Println("a,b:", a, b.Block.ID, b.Block.CollectionID)
		for c, e := range b.Block.ViewIDs {
			//s:=[]string{d}
			//s1 := append(s, "cf7b1a3766ea499a90e568028152b10a")
			if c == 2 {
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
				var query *notionapi.Query = new(notionapi.Query) //=notionapi.Query{[],[],"(WIG","",fliters,[]}
				query.Filter = fliters
				//query.CalendarBy="(WIG"
				query.FilterOperator = "and"
				coll, _ := d.Client.QueryCollection(b.Block.CollectionID, e, query, user)
				//outt,_:=d.Client.GetBlockRecords([]string{e})
				//rr:=outt.Results
				//toVisit := []string{coll.Result.BlockIDS}
				log.Println("rr[0].Block.ViewIDs:", coll.RawJSON, coll.RecordMap.CollectionViews)

				for h, g := range coll.RecordMap.CollectionViews {
					var bb CollectionView
					dd := json.Unmarshal([]byte(g.Value), &bb)
					log.Println("rr[0].Block.ViewIDs:", h, string(g.Value), dd, bb.Page_sort)
					pages=bb.Page_sort
				}
			}
		}
	}
	return pages
}
