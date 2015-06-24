package main

import (
	"net/url"
	"fmt"
	"container/heap"
	"priorityQueue"
	"bufio"
	"os"
	"log"
	"net/http"
	"strings"
	"io/ioutil"
	"github.com/temoto/robotstxt.go"
	"encoding/json"
	"time"
)

var pq *priorityQueue.PriorityQueue = priorityQueue.NewPQueue() //new(priorityQueue.PriorityQueue)
var seenURLS map[string]int = make(map[string]int)
var queuedURLS map[string]int = make(map[string]int)
var domainTimes map[string]int64 = make(map[string]int64)

type LinkStructure struct {
	header string
	RawHtml string
	Text string
	//posns []int
	Link []string
}

func main() {
	fmt.Println("Hello World!")
	heap.Init(pq)

	file, err := os.Open("seedurls")
	if err != nil {
    		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan(){
		obj := priorityQueue.NewURL(strings.TrimSpace(scanner.Text()),0,true)
		obj.Canonicalize();
		queuedURLS[obj.Geturlname()]=0
		heap.Push(pq,obj)
	}
	
	for pq.Len() > 0 && len(seenURLS) <= 10 {
		var temp map[*priorityQueue.URL]bool = make(map[*priorityQueue.URL]bool)
		currUrl := heap.Pop(pq).(*priorityQueue.URL)
		var name string = currUrl.Geturlname()
		delete(queuedURLS,name)
		//fmt.Printf("considering for url %s\n",name)
		hostname,_ := url.Parse(name)
		_,ok := domainTimes[hostname.Host]
		for ok && pq.Len() > 0 {
				//fmt.Printf("did not reach here\n")
				temp[currUrl] = true
				currUrl := heap.Pop(pq).(*priorityQueue.URL)
				var name string = currUrl.Geturlname()
				delete(queuedURLS,name)
				hostname,_ := url.Parse(name)
				_,ok1 := domainTimes[hostname.Host]
				ok=ok1
		}
		if(!ok) {
			//fmt.Printf("not found in timestamps \n")
			ProcessURL(currUrl)
			seenURLS[currUrl.Geturlname()] =0
			domainTimes[hostname.Host] = time.Now().Unix()
		} else {
			temp[currUrl] = true		
			time.Sleep(time.Second*2)
		}
		for key,_ := range temp {
			queuedURLS[key.Geturlname()] = 0
			heap.Push(pq,key)
		}
		removeUnallowedDomains()
	}
}

func removeUnallowedDomains() {
	for key,value := range domainTimes {
		if(time.Now().Unix()-value >= 1) {
			delete(domainTimes, key)
			}
	}
}

func ProcessURL(currUrl *priorityQueue.URL) {
	var name string = currUrl.Geturlname()
	seenURLS[name]=currUrl.GetInlinks()
	if(Isurlok(name) && IsHTMLDoc(name)) {
		CrawlURL(currUrl.Geturlname());
		fmt.Printf("parsed ok")
		} else {
			fmt.Printf("\n%s parsed not ok\n",currUrl.Geturlname())
		}
}

func CrawlURL(uname string) {
	resp, err := http.Get(uname)
	if(err!=nil) {
		fmt.Printf("Error fetching document")
	} else {
		FetchAllOutgoingLinks(uname)
		//addToFrontier(links)
		//extractDocumentInfo(resp.Body)
	}
}

//func extractDocumentInfo(

func FetchAllOutgoingLinks(uname string) {
	getResp,err := http.Get("http://localhost:8080/JSoupRestAPIService/jsoup-api/jsoupapi/"+uname)
	
 if(err!=nil) {
		fmt.Printf("Error in jsoup parsing %s \n",err.Error())
	} else {
		rb, err := ioutil.ReadAll(getResp.Body) 
		//fmt.Printf("%s \n",rb)
	// Check for error
	if err != nil { 
	fmt.Printf("Error in read parsing \n") 
	}	else {
	
	var outlinks []LinkStructure
	json.Unmarshal(rb,&outlinks)
	var fetchedurls map[string]bool = make(map[string]bool)
	
	for i := 0; i< len(outlinks) ; i++ {
		//fmt.Printf("\nurl = %s\n",outlinks[0].Link)
		str := strings.TrimSpace(outlinks[i].Link)
		obj := priorityQueue.NewURL(str,0,false)
		obj.Canonicalize();
		strname := obj.Geturlname()
		count,ok := seenURLS[strname]	
		if !ok {
			count,ok := queuedURLS[strname]
			if !ok {
				_,ok := fetchedurls[strname]
				if(!ok && Isurlok(strname) && IsHTMLDoc(strname)) {
				heap.Push(pq,obj)
				//fmt.Printf("\nadd %s to heap\n",strname)
				fetchedurls[strname] = true
				queuedURLS[strname] = 1
				}
			} else {
				queuedURLS[strname] = count+1
			}
		} else {
			seenURLS[strname] = count+1
		}
	}
	f,err := os.OpenFile("linkGraph.txt", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660);
	if(err!=nil){
		fmt.Printf("Error writing to file %s \n",err)
	} else {
	w := bufio.NewWriter(f)
	w.WriteString(uname+"\t")
	for key,_ := range fetchedurls {
		w.WriteString(key+"\t")
		//fmt.Printf("\nadd %s to heap",key)
	}
	w.WriteString("\n")
	w.Flush()	
	}
	}
	}
}

func Isurlok(str string) bool {
	currurl,err := url.Parse(str);
	if(err==nil) {

	host := strings.ToLower(currurl.Scheme)+"://" +strings.ToLower(currurl.Host) + "/robots.txt"
	resp, err := http.Get(host)
	if(err==nil) {
		robots, err := robotstxt.FromResponse(resp)
		resp.Body.Close()
		if err != nil {
    		fmt.Printf("Error parsing robots.txt:", err.Error())
			return false;
			} else {
				group := robots.FindGroup("*")
			//	fmt.Printf("Error 3 %s \n",group.Test(currurl.Path))
				return group.Test(currurl.Path)
			}
		} else {
		fmt.Printf("Error 1\n")
		return false
		}
		} else {
		fmt.Printf("Error 2\n")
		return false
		}
}

func IsHTMLDoc(str string) bool {
	resp,err := http.Get(str)
	if err!=nil {
		fmt.Printf("\n %s Error checking for html document\n",str)
	} else {
		header := resp.Header.Get("Content-Type")
		//fmt.Printf(resp.)
		if(strings.Contains(header,"text/html")) {
				return true
			} 
		} 
	return false
} 
