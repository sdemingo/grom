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
	"fmt"
	"os"
	"encoding/json"
	"io/ioutil"
	"text/template"
	"sort"
	"strings"
	"image/jpeg"
	"image"
	"errors"
	"io"
)



var months=[]string{
	"Null",
	"Enero",
	"Febrero",
	"Marzo",
	"Abril",
	"Mayo",
	"Junio",
	"Julio",
	"Agosto",
	"Septiembre",
	"Octubre",
	"Noviembre",
	"Diciembre"}



type Articles []*Article


type Blog struct{
	Dir string
	ThemeDir string
	Info BlogInfo
	Posts Articles
	Nposts int
	Statics Articles
	Nstatics int
	Years []bool
	Months []string
}


type BlogInfo map[string]string


type GromConfig map[string]string



func CreateBlog(dir string,themes string)(*Blog,error){

	b:=new (Blog)
	b.Info=make (BlogInfo)
	b.Info["Name"]="My new Blog"
	b.Info["Owner"]="Grom"
	b.Info["Subtitle"]="I wanna be Grom!!"
	b.Info["Theme"]="default"
	b.Info["Url"]="http://yourdomain.com"

	b.Years=make([]bool,100)
	b.Months=months
	b.Nposts=0

	jb, err := json.MarshalIndent(b.Info," "," ")
	if (err!=nil){
		return nil,err
	}

	b.Dir=dir
	b.ThemeDir=b.Dir+"themes/"+b.Info["Theme"]
	os.Mkdir(dir,0755)
	os.Mkdir(dir+"static",0755)
	os.Mkdir(dir+"post",0755)
	os.Mkdir(dir+"img",0755)
	os.Mkdir(dir+"img/thumbs",0755)

	err = createDefaultTheme(dir,themes)
	if err !=nil{
		return nil,err
	}

	err = ioutil.WriteFile(dir+"config.json", jb, 0644)
	if err != nil { 
		return nil,err
	}

	return b,nil
}


func createDefaultTheme(bdir string,tdir string)(error){

	os.Mkdir(bdir+"themes",0755)
	os.Mkdir(bdir+"themes/default",0755)

	fd,err :=os.Open(tdir+"themes/default")
	if err != nil { 
		return errors.New("Grom default theme not found\n")
	}

	names,_:=fd.Readdirnames(-1)
	for i:=range names{
		f1,err:=os.Open(tdir+"themes/default/"+names[i])
		f2,err:=os.Create(bdir+"themes/default/"+names[i])
		io.Copy(f2,f1)
		if err!=nil {
			return err
		}
	}

	return nil
}





func LoadBlog(dir string)(*Blog){

	jb,err := ioutil.ReadFile(dir+"config.json")
	if err != nil { 
		return nil
	}
	
	info:=make (BlogInfo)
	err = json.Unmarshal(jb, &info)

	b:=new (Blog)
	b.Info=info
	b.Dir=dir
	b.ThemeDir=b.Dir+"themes/"+b.Info["Theme"]
	b.Years=make([]bool,100)
	b.Months=months

	b.loadAllPosts()

	b.loadAllStatics()
	return b
}


func (blog *Blog) loadAllPosts()(error){
	fd,err:=os.Open(blog.Dir+"post")
	if err!=nil {
		return err
	}
	posts,_:=fd.Readdirnames(-1)
	blog.Posts=make ([]*Article,len(posts))
	blog.Nposts=0
	for i:=range posts{
		a,_:=ParseArticle(blog.Dir+"post/"+posts[i])
		if a==nil{
			return errors.New("Error parsing "+posts[i])
		}
		if (a!=nil) && (strings.HasSuffix(posts[i],".org")) {
			blog.Posts[i]=a
			if a.Date.Year()>=2000 {
				blog.Years[a.Date.Year()-2000]=true
			}
			blog.Nposts++
		}
		
	}
	sort.Sort(ByDate{blog.Posts})

	return err
}

func (blog *Blog) loadAllStatics()(error){
	fd,err:=os.Open(blog.Dir+"static")
	if err!=nil {
		return err
	}
	statics,_:=fd.Readdirnames(-1)
	blog.Statics=make ([]*Article,len(statics))
	blog.Nstatics=0
	for i:=range statics{
		a,_:=ParseArticle(blog.Dir+"static/"+statics[i])
		if a==nil{
			return errors.New("Error parsing "+statics[i])
		}
		if (a!=nil) && (strings.HasSuffix(statics[i],".org")){
			blog.Statics[i]=a
			blog.Nstatics++
		}
		
	}

	return err
}




func (blog *Blog)AddArticle(title string)(error){

	a,_:=NewArticle(title)
	file:=blog.Dir+"post/"+title+".org"
	err:=a.WriteOrgFile(file)
	if err!=nil{
		return err
	}
	return nil
}


func (blog *Blog)AddStaticPage(title string)(error){

	a,_:=NewArticle(title)
	file:=blog.Dir+"static/"+title+".org"
	err:=a.WriteOrgFile(file)
	if err!=nil{
		return err
	}
	return nil
}


func (blog *Blog)Clean()(error){

	
	return nil
}


func (blog *Blog)Build()(error){
	for i:=range blog.Posts{
		a:=blog.Posts[i]
		if (a!=nil){
			err:=blog.makeArticle(a)
			if err!=nil{
				return err
			}
		}
	}

	for i:=range blog.Statics{
		a:=blog.Statics[i]
		if (a!=nil){
			err:=blog.makeStatic(a)
			if err!=nil{
				return err
			}
		}
	}
	
	err:=blog.makeIndex()
	if err!=nil{
		return err
	}

	err=blog.makeArchive()
	if err!=nil{
		return err
	}

	err=blog.makeThumbs()
	if err!=nil{
		return err
	}

	err=blog.makeSitemap()
	if err!=nil{
		return err
	}

	return nil
}



/*
 Mecanismo de ordenación de los arrays de artículos
*/
func (s Articles) Len() int      { return len(s) }
func (s Articles) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByDate struct{ Articles }
func (s ByDate) Less(i, j int) bool { 
	a:=s.Articles[i]
	b:=s.Articles[j]
	if (a==nil) || (b==nil){
		return false
	}
	return s.Articles[i].Date.After(s.Articles[j].Date)
}







type BlogDataTemplate struct{
	Blog *Blog
	Pages []*Article
}


func (blog *Blog)GetArticleId(a* Article)(string){
	if a==nil {
		return ""
	}
	d:=a.GetDate()
	ds:=d.Format("2006-01")
	return ds+"-"+a.Id
}


func (blog *Blog)GetArticlesByDate(year int,month int)([]*Article){

	a:=make([]*Article,100)

	n:=0
	date:=""
	year=year+2000
	if (month<10){
		date=fmt.Sprintf("%d-0%d",year,month)
	}else{
		date=fmt.Sprintf("%d-%d",year,month)
	}

	for p:=0;p<blog.Nposts;p++{
		if strings.HasPrefix(blog.GetArticleId(blog.Posts[p]), date){
			a=append(a,blog.Posts[p])
			n++
		}
	}

	if (n==0){
		return nil
	}
	return a
}






func (blog *Blog)makeIndex()(error){

	f,err:=os.Create(blog.Dir+"index.html")
	if err!=nil{
		return err
	}

	n_post_in_index:=len(blog.Posts)
	if n_post_in_index>5 {
		n_post_in_index=5  //max 5 articles in index
	}

	data:=new (BlogDataTemplate)
	data.Blog=blog
	data.Pages=blog.Posts[:n_post_in_index]

	t:=template.New("main")
	_,err=t.ParseFiles(blog.ThemeDir+"/main.html",
		blog.ThemeDir+"/post.html")
	if (err!=nil){
		return err
	}

	err=t.ExecuteTemplate(f,"main",data)
	if (err!=nil){
		return err
	}
	return nil
}




func (blog *Blog) makeStatic(a *Article)(error){
	
	f,err:=os.Create(blog.Dir+"static-"+a.Id+".html")
	if err!=nil{
		return err
	}

	s:=-1

	// search the index of the article
	for i:=range blog.Statics{
		if  a.Id==blog.Statics[i].Id {
			s=i
			break
		}
	}

	data:=new (BlogDataTemplate)
	data.Blog=blog
	data.Pages=blog.Statics[s:s+1]

	t:=template.New("main")
	_,err=t.ParseFiles(blog.ThemeDir+"/main.html",
		blog.ThemeDir+"/static.html")
	if (err!=nil){
		return err
	}

	t.ExecuteTemplate(f,"main",data)

	return nil
}



func (blog *Blog) makeArticle(a *Article)(error){
	
	f,err:=os.Create(blog.Dir+blog.GetArticleId(a)+".html")
	if err!=nil{
		return err
	}

	s:=-1

	// search the index of the article
	for i:=range blog.Posts{
		if  ((blog.Posts[i]!=nil) && (a.Id==blog.Posts[i].Id)) {
			s=i
			break
		}
	}

	if s<0{
		return errors.New("Bad article to build")
	}

	data:=new (BlogDataTemplate)
	data.Blog=blog
	data.Pages=blog.Posts[s:s+1]

	t:=template.New("main")
	_,err=t.ParseFiles(blog.ThemeDir+"/main.html",
		blog.ThemeDir+"/post.html")
	if (err!=nil){
		return err
	}

	t.ExecuteTemplate(f,"main",data)

	return nil
}



func (blog *Blog)makeArchive()(error){

	f,err:=os.Create(blog.Dir+"archive.html")
	if err!=nil{
		return err
	}

	data:=new (BlogDataTemplate)
	data.Blog=blog
	data.Pages=blog.Posts

	t:=template.New("main")
	_,err=t.ParseFiles(blog.ThemeDir+"/main.html",
		blog.ThemeDir+"/archive.html")
	if (err!=nil){
		return err
	}

	err=t.ExecuteTemplate(f,"main",data)
	if (err!=nil){
		return err
	}

	return nil
}


func (blog *Blog)makeThumbs()(error){

	fd,err:=os.Open(blog.Dir+"img")
	if err!=nil{
		return err
	}
	imgs,_:=fd.Readdirnames(-1)
	for i:=range imgs{
		if imgs[i]!="thumbs"{
			blog.createThumb(imgs[i])
		}
	}

	return nil
}


func (blog *Blog)createThumb(file string)(error){

	var img1 image.Image
	var delta float32

	fimg, err := os.Open(blog.Dir+"img/"+file)	
	img1, err = jpeg.Decode(fimg)
	if err!=nil{
		return err
	}

	r:=img1.Bounds()
	s:=r.Size()
	
	if (s.X > 500){
		delta=float32(s.X)/500.0
	}else{
		delta=1.0
	}

	nx:=float32(s.X)/delta
	ny:=float32(s.Y)/delta
	
	img2:=Resize(img1,r,int(nx),int(ny))

	fimg2,_:=os.Create(blog.Dir+"img/thumbs/"+file)
	jpeg.Encode(fimg2,img2,&jpeg.Options{jpeg.DefaultQuality})

	return nil
}



/*
 Sitemap generator
*/

var sitemapTemplate=`{{define "sitemap"}}<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
{{$b:=.}}
{{ range $a:=.Posts}}
  <url>
      <loc>{{$b.Info.Url}}/{{$b.GetArticleId $a}}.html</loc>
      <lastmod>{{$a.GetStringSitemapDate}}</lastmod>
      <changefreq>monthly</changefreq>
      <priority>0.8</priority>
   </url>
{{end}}
{{ range $a:=.Statics}}
  <url>
      <loc>{{$b.Info.Url}}/{{$b.GetArticleId $a}}.html</loc>
      <lastmod>{{$a.GetStringSitemapDate}}</lastmod>
      <changefreq>monthly</changefreq>
      <priority>0.8</priority>
   </url>
</urlset>
{{end}}
{{end}}
`

func (blog *Blog)makeSitemap()(error){
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
