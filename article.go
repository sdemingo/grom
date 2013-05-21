/**

   Grom

   Copyright 2013 Sergio de Mingo

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

*/

package main

import (
	"time"
	"regexp"
	"strings"
	"io/ioutil"
	"errors"
)


type Article struct{
	Id string
	Title string
	Date time.Time
	DateFormat map[string]string
	Content []byte
	Meta map[string]string

}


var checkID = regexp.MustCompile("[^(\\w|\\.)]")


const(
	RSSDateFormat  = "Mon, 02 Jan 2006 15:04:05 GMT"
	AtomDateFormat = time.RFC3339
	OrgDateFormat  = "2006-01-02"
	PostDateFormat = "Mon, 02 Jan 2006"
	
)

func NewArticle(id string)(*Article,error){
	a:=new(Article)
	a.Date=time.Now()
	a.DateFormat=make(map[string]string)
	a.DateFormat["PostDateFormat"]=a.GetDateString(PostDateFormat)
	a.DateFormat["SitemapDateFormat"]=a.GetDateString(OrgDateFormat)
	a.DateFormat["RSSDateFormat"]=a.GetDateString(RSSDateFormat)

	a.Id=checkID.ReplaceAllString(id,"-")


	a.Meta=make(map[string]string)
	a.Meta["Id"]=a.Id
	a.Meta["Date"]=date2String(a.Date)
	a.Meta["Author"]="Sergio de Mingo"

	return a,nil
}

func ParseArticle(ifile string)(*Article,error){
	a:=new(Article)
	
	b, err := ioutil.ReadFile(ifile)
	if err != nil { 
		return nil,err
	}

	a.Content=b
	a.Title=parseTitle(a.Content)
	a.Meta=make(map[string]string)
	a.Meta["Id"]=parseProperty(a.Content,"Id")
	a.Meta["Date"]=parseProperty(a.Content,"Date")
	a.Meta["Author"]=parseProperty(a.Content,"Author")
	a.Date,err=parseDate(a.Meta["Date"])
	if err!=nil {
		return nil,errors.New("Article with corrupted date")
	}
	a.Id=a.Meta["Id"]
	
	a.Date,_=parseDate(a.Meta["Date"])
	a.DateFormat=make(map[string]string)
	a.DateFormat["PostDateFormat"]=a.GetDateString(PostDateFormat)
	a.DateFormat["SitemapDateFormat"]=a.GetDateString(OrgDateFormat)
	a.DateFormat["RSSDateFormat"]=a.GetDateString(RSSDateFormat)
	
	return a,nil
}


func (a *Article) GetHTMLContent()(string){
	return string(convertToHtml(a.Content))
}

/*
func (a *Article) GetStringContent()(string){
	return string(a.Content)
}
*/


func (a *Article)GetValidId()(string){
	if a==nil {
		return ""
	}
	d:=a.GetDate()
	ds:=d.Format("2006-01")
	return ds+"-"+a.Id
}



func (a *Article) GetDate()(time.Time){
	t,_:=parseDate(a.Meta["Date"])
	return t
}


func (a* Article) GetDateString(format string)(string){
	return a.Date.Format(format)
}



func (a *Article) WriteOrgFile(ofile string)(error){
	s:="* Article Title\n"
	s=s+":PROPERTIES:\n"
	for k,v:=range a.Meta{
		s=s+":"+k+": "+v+"\n"
	}
	s=s+":END:\n"
	s=s+"\n\n Write your article!\n\n"
	
	err := ioutil.WriteFile(ofile, []byte(s), 0644)
	if err != nil { 
		return err
	}
	
	return nil	
}




/*
       Private Methods
*/


func date2String(date time.Time)(string){
	s:=date.Format("<"+OrgDateFormat)
	s=s+" "+date.Weekday().String()[:3]+">"
	return s
}


func parseDate(orgdate string)(time.Time,error){
	
	dayReg:= regexp.MustCompile("[a-zA-Z\\>\\< ]+")
	orgdate=dayReg.ReplaceAllString(orgdate,"")
	t,err:=time.Parse("2006-01-02",orgdate)
	return t,err
}


func parseProperty(content []byte, key string)(string){

	propReg:= regexp.MustCompile("(?m)^:"+key+":.+$")
	p:=string(propReg.Find(content))
	f:=strings.Split(p,":"+key+":")
	if ((f==nil) || (len(f)<2)){
		return ""
	}
	return strings.Trim(f[1]," \t")
}


func parseTitle(content []byte)(string){

	propReg:= regexp.MustCompile("(?m)^\\* .+$")
	p:=string(propReg.Find(content))
	f:=strings.Split(p,"*")
	if ((f==nil) || (len(f)<2)){
		return ""
	}
	return strings.Trim(f[1]," \t")
}



/*
 HTML conversion
*/

var head1Reg = regexp.MustCompile("(?m)^\\* (?P<head>.+)\\n")
var head2Reg = regexp.MustCompile("(?m)^\\*\\* (?P<head>.+)\\n")
var linkReg = regexp.MustCompile("\\[\\[(?P<url>[^\\]]+)\\]\\[(?P<text>[^\\]]+)\\]\\]")
var imgLinkReg = regexp.MustCompile("\\[\\[file:\\.\\./img/(?P<img>[^\\]]+)\\]\\[file:\\.\\./img/(?P<thumb>[^\\]]+)\\]\\]")
var imgReg = regexp.MustCompile("\\[\\[\\.\\./img/(?P<src>[^\\]]+)\\]\\]")
var codeReg = regexp.MustCompile("(?m)^\\#\\+BEGIN_SRC \\w*\\n(?P<code>(?s).+)^\\#\\+END_SRC\\n")
var quoteReg = regexp.MustCompile("(?m)^\\#\\+BEGIN_QUOTE\\s*\\n(?P<cite>(?s).+)^\\#\\+END_QUOTE\\n")
var parReg = regexp.MustCompile("\\n\\n+(?P<text>[^\\n]+)")
var allPropsReg = regexp.MustCompile(":PROPERTIES:(?s).+:END:")
var rawHTML = regexp.MustCompile("\\<[^\\>]+\\>")

//estilos de texto
var boldReg = regexp.MustCompile("(?P<prefix>[\\s|\\W]+)\\*(?P<text>[^\\s][^\\*]+)\\*(?P<suffix>[\\s|\\W]*)")
var italicReg = regexp.MustCompile("(?P<prefix>[\\s])/(?P<text>[^\\s][^/]+)/(?P<suffix>[^A-Za-z0-9]*)")
var ulineReg = regexp.MustCompile("(?P<prefix>[\\s|\\W]+)_(?P<text>[^\\s][^_]+)_(?P<suffix>[\\s|\\W]*)")
var codeLineReg = regexp.MustCompile("(?P<prefix>[\\s|\\W]+)=(?P<text>[^\\s][^\\=]+)=(?P<suffix>[\\s|\\W]*)")
var strikeReg = regexp.MustCompile("(?P<prefix>[\\s|[\\W]+)\\+(?P<text>[^\\s][^\\+]+)\\+(?P<suffix>[\\s|\\W]*)")


// listas
var ulistItemReg = regexp.MustCompile("(?m)^\\s*[\\+|\\-]\\s*(?P<item>.+)\\n")
var olistItemReg = regexp.MustCompile("(?m)^\\s*[0-9]+\\.\\s*(?P<item>.+)\\n")
var ulistReg = regexp.MustCompile("(?P<items>(\\<fake-uli\\>.+\\n)+)")
var olistReg = regexp.MustCompile("(?P<items>(\\<fake-oli\\>.+\\n)+)")



func convertToHtml(content []byte)([]byte){
	// First remove all HTML raw tags for security
	out:=rawHTML.ReplaceAll(content,[]byte(""))

	// headings (h1 is not admit in the post body)
	out=head1Reg.ReplaceAll(out,[]byte(""))
	out=head2Reg.ReplaceAll(out,[]byte("<h2>$head</h2>\n"))


	// images and blocks
	out=imgReg.ReplaceAll(out,[]byte("<div class='image'><a href='img/$src'><img src='img/thumbs/$src'/></a></div>"))
	out=imgLinkReg.ReplaceAll(out,[]byte("<div class='image'><a href='img/$img'><img src='img/thumbs/$thumb'/></a></div>"))
	out=linkReg.ReplaceAll(out,[]byte("<a href='$url'>$text</a>"))
	out=codeReg.ReplaceAll(out,[]byte("<pre><code>$code</code></pre>\n"))
	out=quoteReg.ReplaceAll(out,[]byte("<blockquote>$cite</blockquote>\n"))
	//out=parReg.ReplaceAll(out,[]byte(".\n<p>"))
	out=parReg.ReplaceAll(out,[]byte("\n\n<p/>$text"))
	out=allPropsReg.ReplaceAll(out,[]byte("\n"))


	// font styles

	out=italicReg.ReplaceAll(out,[]byte("$prefix<i>$text</i>$suffix"))
	out=boldReg.ReplaceAll(out,[]byte("$prefix<b>$text</b>$suffix"))
	out=ulineReg.ReplaceAll(out,[]byte("$prefix<u>$text</u>$suffix"))
	out=codeLineReg.ReplaceAll(out,[]byte("$prefix<code>$text</code>$suffix"))
	out=strikeReg.ReplaceAll(out,[]byte("$prefix<s>$text</s>$suffix"))


	// List with fake tags for items
	out=ulistItemReg.ReplaceAll(out,[]byte("<fake-uli>$item</fake-uli>\n"))
	out=ulistReg.ReplaceAll(out,[]byte("<ul>\n$items</ul>\n"))
	out=olistItemReg.ReplaceAll(out,[]byte("<fake-oli>$item</fake-oli>\n"))
	out=olistReg.ReplaceAll(out,[]byte("<ol>\n$items</ol>\n"))

	// Removing fake items tags
	sout:=string(out)
	sout=strings.Replace(sout,"<fake-uli>","<li>",-1)
	sout=strings.Replace(sout,"</fake-uli>","</li>",-1)
	sout=strings.Replace(sout,"<fake-oli>","<li>",-1)
	sout=strings.Replace(sout,"</fake-oli>","</li>",-1)
	
	return []byte(sout)
}


