// URL
package priorityQueue

import (
	"strings"
	"net/url"
	"fmt"
)

type URL struct {
	urlname string 
	in_links int
	seed bool
	index int
}

func NewURL(name string,inlinks int,s bool) *URL {
    return &URL{
		urlname: name,
		in_links: inlinks,
		seed: s,
	}
}

func (url *URL) Geturlname() string {
	return url.urlname
}

func (url *URL) GetInlinks() int {
	return url.in_links
}

func (obj *URL) Canonicalize() {
	obj.urlname = modifyURL(obj.urlname);
}

func modifyURL(s string) string {
	strings.Replace(s,":80","",1)
	strings.Replace(s,":443","",1)
	
	init,err := url.Parse(s)
	parsed := init.ResolveReference(init)
	
	if(err!=nil) {fmt.Print("error in parsing url")}
	
	var scheme string
	scheme = strings.ToLower(parsed.Scheme)
	
	if scheme==""&&strings.HasPrefix(s,"//") {
		scheme = "http"
	}
	
	host := strings.ToLower(parsed.Host)
	path := parsed.Path
	query := parsed.RawQuery
	
	if query == "" { 
		return (scheme+"://"+host+path) 
	}
	return scheme+"://"+host+path+"?"+query
}