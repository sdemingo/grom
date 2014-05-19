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
	//"fmt"
	"os"
	"text/template"
	"time"
	"regexp"
	"strings"
)


const (
	POPULAR_TAGS_TO_SHOW=4
)

type Tag struct{
	Name string
	Posts Articles
	Nposts int
}

type Tags map[string] Tag

type TagsSlice [] Tag

func (t TagsSlice) Len() int { return len(t) }
func (t TagsSlice) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t TagsSlice) Less(i, j int) bool {return t[i].Nposts < t[j].Nposts }


func (t *Tag) getValidId()(string){
	return normalizeURL(t.Name)
}


func normalizeURL(url string)(string){
	urlReg:=regexp.MustCompile("[^a-zA-Z0-9áéíóúÁÉÍÓÚñÑ\\-_]+")
	if urlReg.MatchString(url) {
		return ""
	}
	s:=url
	s=strings.Replace(s,"ñ","n",-1)
	s=strings.Replace(s,"á","a",-1)
	s=strings.Replace(s,"é","e",-1)
	s=strings.Replace(s,"í","i",-1)
	s=strings.Replace(s,"ó","o",-1)
	s=strings.Replace(s,"ú","u",-1)
	return s
}


func (t Tag)makeTagIndex(blog *Blog)(error){
	f,err:=os.Create(blog.Dir+"/tags/"+t.getValidId()+".html")
	if err!=nil{
		return err
	}

	blog.TagSelected=t

	tmpl:=template.New("main")
	_,err=tmpl.ParseFiles(blog.ThemeDir+"/main.html",
		blog.ThemeDir+"/tag-index.html")
	if (err!=nil){
		return err
	}

	err=tmpl.ExecuteTemplate(f,"main",blog)
	if (err!=nil){
		return err
	}

	return nil
}


func makeAllTagsIndex(blog *Blog)(error){
	f,err:=os.Create(blog.Dir+"/tags/index.html")
	if err!=nil{
		return err
	}

	tmpl:=template.New("main")
	_,err=tmpl.ParseFiles(blog.ThemeDir+"/main.html",
		blog.ThemeDir+"/all-tags.html")
	if (err!=nil){
		return err
	}

	err=tmpl.ExecuteTemplate(f,"main",blog)
	if (err!=nil){
		return err
	}

	return nil
}


func buildTags(blog *Blog)(error){

	blog.BlogTags=make(Tags)

	for i:=range blog.Posts{
		a:=blog.Posts[i]
		if (a!=nil){
			names:=strings.Split(blog.Posts[i].Meta["Tags"],",")
			for t:=0;t<len(names);t++{
				var tag Tag
				var ok bool
				name:=strings.Trim(names[t]," ")
				n_name:=normalizeURL(name)
				if ((name=="") || (n_name=="")){
					continue
				}
				if tag,ok=blog.BlogTags[n_name];!ok {
					tag.Name=name
					tag.Posts=make([]*Article,0)	
					tag.Nposts=0
				}
				tag.Posts=append(tag.Posts,blog.Posts[i])
				tag.Nposts++
				blog.BlogTags[n_name]=tag
				a.ArticleTags[n_name]=tag
			}
		}
	}

	for _,v:=range blog.BlogTags{
		err:=v.makeTagIndex(blog)
		if err!=nil{
			return err
		}
	}

	makeAllTagsIndex(blog)

	return nil
}





var sitemapTemplate=`{{define "sitemap"}}<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
{{$b:=.}}
{{ range $a:=.Posts}}
{{ if $a}}
  <url>
      <loc>{{$b.Info.Url}}/html/{{$a.GetYear}}/{{$a.GetValidId}}.html</loc>
      <lastmod>{{$a.DateFormat.SitemapDateFormat}}</lastmod>
      <changefreq>monthly</changefreq>
      <priority>0.8</priority>
   </url>
{{end}}
{{end}}
{{ range $a:=.Statics}}
  <url>
      <loc>{{$b.Info.Url}}/html/{{$a.GetValidId}}.html</loc>
      <lastmod>{{$a.DateFormat.SitemapDateFormat}}</lastmod>
      <changefreq>monthly</changefreq>
      <priority>0.8</priority>
   </url>
</urlset>
{{end}}
{{end}}
`



var atomTemplate=`{{define "atom"}}<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
<id>{{.Info.Url}}/atom.xml</id>
<title>{{.Info.Name}}</title>
<subtitle>{{.Info.Subtitle}}</subtitle>
<link href="{{.Info.Url}}/atom.xml" rel="self" />
<link href="{{.Info.Url}}" />
<updated>{{.GetFeedDate}}</updated>
{{$b:=.}}
{{ range $a:=.GetLastArticles}}

<entry>
<title>{{$a.Title}}</title>
<link href="{{$b.Info.Url}}/html/{{$a.GetYear}}/{{$b.GetArticleId $a}}.html" />
<id>{{$b.Info.Url}}/html/{{$a.GetValidId}}.html</id>
<updated>{{$a.DateFormat.AtomDateFormat}}</updated>
<author>
<name>{{$a.Meta.Author}}</name>
</author>
<description><![CDATA[{{$b.GetHTMLContent $a}}]]></description>
</entry>

{{end}}
</feed>
{{end}}
`

var rssTemplate=`{{define "rss"}}<?xml version="1.0" encoding="utf-8" ?>
<rss version="2.0">
<channel>
<title>{{.Info.Name}}</title>
<link>{{.Info.Url}}</link>
<description>{{.Info.Subtitle}}</description>
{{$b:=.}}
{{ range $a:=.GetLastArticles}}
<item>
<title>{{$a.Title}}</title>
<pubDate>{{$a.DateFormat.RSSDateFormat}}</pubDate>
<guid>{{$b.Info.Url}}/html/{{$a.GetYear}}/{{$a.GetValidId}}.html</guid>
<link>{{$b.Info.Url}}/html/{{$a.GetYear}}/{{$a.GetValidId}}.html</link>
<description><![CDATA[{{$b.GetHTMLContent $a}}]]></description>
</item>
{{end}}
</channel>
</rss>
{{end}}
`




func makeSitemap(blog *Blog)(error){
	f,err:=os.Create(blog.Dir+"sitemap.xml")
	if err!=nil{
		return err
	}

	t:=template.New("sitemap")
	_,err=t.Parse(sitemapTemplate)
	if (err!=nil){
		return err
	}

	err=t.ExecuteTemplate(f,"sitemap",blog)
	if (err!=nil){
		return err
	}

	return nil
}




func getFeedDate()(string){
	t:=time.Now()
	return t.Format(AtomDateFormat)
}


func makeAtomFeed(blog *Blog)(error){
	f,err:=os.Create(blog.Dir+"atom.xml")
	if err!=nil{
		return err
	}

	t:=template.New("atom")
	_,err=t.Parse(atomTemplate)
	if (err!=nil){
		return err
	}

	err=t.ExecuteTemplate(f,"atom",blog)
	if (err!=nil){
		return err
	}

	return nil
}


func makeRSSFeed(blog *Blog)(error){
	f,err:=os.Create(blog.Dir+"rss.xml")
	if err!=nil{
		return err
	}

	t:=template.New("rss")
	_,err=t.Parse(rssTemplate)
	if (err!=nil){
		return err
	}

	err=t.ExecuteTemplate(f,"rss",blog)
	if (err!=nil){
		return err
	}

	return nil
}



