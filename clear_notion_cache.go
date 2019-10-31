package main

import (
	"github.com/kjk/notionapi/caching_downloader"
	"log"
)

func clear(d *caching_downloader.Downloader)  {
	pages,_:=d.Cache.GetPageIDs()
	records,err:=d.Client.GetBlockRecords(pages)
	if err!=nil{
		log.Println(err)
		//return false
	}
	for a,record :=range records.Results{
		if record.Role=="none"{
			d.Cache.Remove(pages[a]+".txt")
			log.Println("已清理record.ID:",pages[a])
		}
	}
}