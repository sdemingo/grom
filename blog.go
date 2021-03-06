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
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var months = []string{
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

type Blog struct {
	Dir         string
	ThemeDir    string
	Info        BlogInfo
	Posts       Articles
	Nposts      int
	Statics     Articles
	Nstatics    int
	Years       []bool
	Months      []string
	Selected    int
	BlogTags    Tags //all tags
	TagSelected Tag
	DebugServer *WebSockServer
}

type BlogInfo map[string]string

func CreateBlog(dir string, themes string) (*Blog, error) {

	b := new(Blog)
	b.Info = make(BlogInfo)
	b.Info["Name"] = "My new Blog"
	b.Info["Owner"] = "Grom"
	b.Info["Subtitle"] = "I wanna be Grom!!"
	b.Info["Theme"] = "default"
	b.Info["Url"] = "http://yourdomain.com"

	b.Years = make([]bool, 100)
	b.Months = months
	b.Posts = make([]*Article, 500)
	b.Nposts = 0

	jb, err := json.MarshalIndent(b.Info, " ", " ")
	if err != nil {
		return nil, err
	}

	b.Dir = dir
	b.ThemeDir = b.Dir + "themes/" + b.Info["Theme"]
	os.Mkdir(dir, 0755)
	os.Mkdir(dir+"static", 0755)
	os.Mkdir(dir+"post", 0755)
	os.Mkdir(dir+"html", 0755)
	os.Mkdir(dir+"img", 0755)
	os.Mkdir(dir+"img/thumbs", 0755)

	err = createDefaultTheme(dir, themes)
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(dir+"config.json", jb, 0644)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func createDefaultTheme(bdir string, tdir string) error {

	os.Mkdir(bdir+"themes", 0755)
	os.Mkdir(bdir+"themes/default", 0755)

	fd, err := os.Open(tdir + "themes/default")
	if err != nil {
		return errors.New("Grom default theme not found\n")
	}

	names, _ := fd.Readdirnames(-1)
	for i := range names {
		f1, err := os.Open(tdir + "themes/default/" + names[i])
		f2, err := os.Create(bdir + "themes/default/" + names[i])
		io.Copy(f2, f1)
		if err != nil {
			return err
		}
	}

	return nil
}

func LoadBlog(dir string) *Blog {

	jb, err := ioutil.ReadFile(dir + "config.json")
	if err != nil {
		fmt.Println(err)
		return nil
	}

	info := make(BlogInfo)
	err = json.Unmarshal(jb, &info)

	b := new(Blog)
	b.Info = info
	b.Dir = dir
	b.ThemeDir = b.Dir + "themes/" + b.Info["Theme"]
	b.Years = make([]bool, 100)
	b.Months = months

	b.Posts = make([]*Article, 500)
	b.Nposts = 0

	b.loadAllPosts()
	b.loadAllStatics()
	return b
}

func (blog *Blog) loadFilePost(fp string, fi os.FileInfo, err error) error {
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if fi.IsDir() {
		return nil
	}

	//if !strings.HasSuffix(fp, ".org") {
	//	return nil // file is not org
	//
	a, err := ParseArticle(fp)
	if a == nil {
		fmt.Println("Error parsing " + err.Error())
		return nil
	}

	// Array limits not controlled
	blog.Posts[blog.Nposts] = a
	blog.Nposts++
	if a.Date.Year() >= 2000 {
		blog.Years[a.Date.Year()-2000] = true
	}

	return nil
}

func (blog *Blog) loadAllPosts() {
	err := filepath.Walk(blog.Dir+"post", blog.loadFilePost)
	if err != nil {
		fmt.Println(err)
	}
	blog.Posts = blog.Posts[:blog.Nposts]
	sort.Sort(ByDate{blog.Posts})

}

func (blog *Blog) loadAllStatics() error {
	fd, err := os.Open(blog.Dir + "static")
	if err != nil {
		return err
	}
	statics, _ := fd.Readdirnames(-1)
	blog.Statics = make([]*Article, len(statics))
	blog.Nstatics = 0
	for i := range statics {
		a, _ := ParseArticle(blog.Dir + "static/" + statics[i])
		if a == nil {
			return errors.New("Error parsing " + statics[i])
		}
		if a != nil {
			/*
			 Array limits not controlled
			*/
			blog.Statics[i] = a
			blog.Nstatics++
		}

	}

	return err
}

func (blog *Blog) AddArticle(title string) error {

	d := time.Now()
	year := d.Format("2006")
	month := d.Format("01")

	a, _ := NewArticle(title)
	file := blog.Dir + "post/" + year + "/" + month + "-" + title + ".md"
	err := a.WriteNewFile(file)
	if err != nil {
		return err
	}
	return nil
}

func (blog *Blog) AddStaticPage(title string) error {

	a, _ := NewArticle(title)
	file := blog.Dir + "static/" + title + ".md"
	err := a.WriteNewFile(file)
	if err != nil {
		return err
	}
	return nil
}

func (blog *Blog) Serve() error {
	//http.HandleFunc("/ws", blog.DebugServer.WSHandler)
	//fmt.Println("Añado el handler del websocket")

	http.ListenAndServe(":9999", http.FileServer(http.Dir(blog.Dir)))
	return nil
}

func (blog *Blog) Build() error {

	fmt.Printf("Building tags ... ")
	err := blog.makeTags()
	if err != nil {
		return err
	}
	fmt.Printf("\n")

	fmt.Printf("Building posts ... ")
	for i := range blog.Posts {
		a := blog.Posts[i]
		if a != nil {
			err := blog.makeArticle(a)
			if err != nil {
				return err
			}
		}
	}
	fmt.Printf("\n")

	fmt.Printf("Building statics ... ")
	for i := range blog.Statics {
		a := blog.Statics[i]
		if a != nil {
			err := blog.makeStatic(a)
			if err != nil {
				return err
			}
		}
	}
	fmt.Printf("\n")

	fmt.Printf("Building index ... ")
	err = blog.makeIndex()
	if err != nil {
		return err
	}
	fmt.Printf("\n")

	fmt.Printf("Building archive ... ")
	err = blog.makeArchive()
	if err != nil {
		return err
	}
	fmt.Printf("\n")

	fmt.Printf("Building images and thumbs ... ")
	err = blog.makeThumbs()
	if err != nil {
		return err
	}
	fmt.Printf("\n")

	fmt.Printf("Building blog utils ... ")
	err = blog.BuildUtils()
	if err != nil {
		return err
	}
	fmt.Printf("\n")

	return nil
}

func (blog *Blog) BuildUtils() error {

	err := makeSitemap(blog)
	if err != nil {
		return err
	}

	err = makeRSSFeed(blog)
	if err != nil {
		return err
	}

	return nil
}

func (blog *Blog) GetArticlesByDate(year int, month int) []*Article {

	a := make([]*Article, 100)
	n := 0
	smonth := ""
	year = year + 2000
	syear := fmt.Sprintf("%d", year)
	if month < 10 {
		smonth = fmt.Sprintf("0%d", month)
	} else {
		smonth = fmt.Sprintf("%d", month)
	}

	for p := 0; p < blog.Nposts; p++ {
		if (blog.Posts[p].GetYear() == syear) &&
			strings.HasPrefix(blog.Posts[p].GetValidId(), smonth) {

			a = append(a, blog.Posts[p])
			n++
		}
	}

	if n == 0 {
		return nil
	}
	return a
}

func (blog *Blog) GetLastArticles() []*Article {

	var max string
	var ok bool

	if max, ok = blog.Info["PostPerPage"]; !ok {
		panic(errors.New("No PostPerPage defined in config.json"))
	}

	max_posts, err := strconv.ParseInt(max, 10, 0)
	if err != nil {
		panic(errors.New("Bad value for PostPerPage defined in config.json"))
	}

	n_post_in_index := len(blog.Posts)
	if n_post_in_index > int(max_posts) {
		n_post_in_index = int(max_posts)
	}

	return blog.Posts[:n_post_in_index]
}

func (blog *Blog) GetSelectedPost() *Article {
	return blog.Posts[blog.Selected]
}

func (blog *Blog) GetSelectedStatic() *Article {
	return blog.Statics[blog.Selected]
}

func (blog *Blog) GetHTMLContent(a *Article) string {
	return Markdown2HTML(a.Content, blog.Info["Url"])
}

func (blog *Blog) GetPopularTags() Tags {

	popular := make(TagsSlice, 0, len(blog.BlogTags))
	for _, v := range blog.BlogTags {
		popular = append(popular, v)
	}
	sort.Sort(sort.Reverse(popular))
	popular = popular[:POPULAR_TAGS_TO_SHOW]

	popularMap := make(Tags)
	for i := range popular {
		n_name := normalizeURL(popular[i].Name)
		popularMap[n_name] = popular[i]
	}

	return popularMap
}

/*
 Mecanismo de ordenación de los arrays de artículos
*/
func (s Articles) Len() int      { return len(s) }
func (s Articles) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByDate struct{ Articles }

func (s ByDate) Less(i, j int) bool {
	a := s.Articles[i]
	b := s.Articles[j]
	if (a == nil) || (b == nil) {
		return false
	}
	return s.Articles[i].Date.After(s.Articles[j].Date)
}

func (blog *Blog) makeIndex() error {

	f, err := os.Create(blog.Dir + "index.html")
	if err != nil {
		return err
	}

	t := template.New("main")
	_, err = t.ParseFiles(blog.ThemeDir+"/main.html",
		blog.ThemeDir+"/last-posts.html")
	if err != nil {
		return err
	}

	err = t.ExecuteTemplate(f, "main", blog)
	if err != nil {
		return err
	}
	return nil
}

func (blog *Blog) makeStatic(a *Article) error {
	f, err := os.Create(blog.Dir + "html/static-" + a.Id + ".html")
	if err != nil {
		return err
	}
	s := -1
	// search the index of the article
	for i := range blog.Statics {
		if a.Id == blog.Statics[i].Id {
			s = i
			break
		}
	}
	if s < 0 {
		return errors.New("Bad static to build")
	}

	blog.Selected = s
	t := template.New("main")
	_, err = t.ParseFiles(blog.ThemeDir+"/main.html",
		blog.ThemeDir+"/static.html")
	if err != nil {
		return err
	}
	t.ExecuteTemplate(f, "main", blog)

	return nil
}

func (blog *Blog) makeArticle(a *Article) error {

	err := os.MkdirAll(blog.Dir+"html/"+a.GetYear(), 0755)
	if err != nil {
		return err
	}
	f, err := os.Create(blog.Dir + "html/" + a.GetYear() + "/" + a.GetValidId() + ".html")
	if err != nil {
		return err
	}
	s := -1
	// search the index of the article
	for i := range blog.Posts {
		if (blog.Posts[i] != nil) && (a.Id == blog.Posts[i].Id) {
			s = i
			break
		}
	}
	if s < 0 {
		return errors.New("Bad article to build")
	}

	blog.Selected = s
	t := template.New("main")
	_, err = t.ParseFiles(blog.ThemeDir+"/main.html",
		blog.ThemeDir+"/post.html")
	if err != nil {
		return err
	}
	t.ExecuteTemplate(f, "main", blog)

	return nil
}

func (blog *Blog) makeTags() error {

	return buildTags(blog)
}

func (blog *Blog) makeArchive() error {

	f, err := os.Create(blog.Dir + "html/archive.html")
	if err != nil {
		return err
	}

	t := template.New("main")
	_, err = t.ParseFiles(blog.ThemeDir+"/main.html",
		blog.ThemeDir+"/archive.html")
	if err != nil {
		return err
	}

	err = t.ExecuteTemplate(f, "main", blog)
	if err != nil {
		return err
	}

	return nil
}

func (blog *Blog) makeThumbs() error {

	fd, err := os.Open(blog.Dir + "img")
	if err != nil {
		return err
	}
	imgs, _ := fd.Readdirnames(-1)
	for i := range imgs {
		if imgs[i] != "thumbs" {
			blog.createThumb(imgs[i])
		}
	}

	return nil
}

func (blog *Blog) createThumb(file string) error {

	var img1 image.Image
	var delta float32

	_, err := os.Stat(blog.Dir + "img/thumbs/" + file)
	if err == nil {
		return nil // the thumb exits. exit
	}
	fimg, err := os.Open(blog.Dir + "img/" + file)
	img1, err = jpeg.Decode(fimg)
	if err != nil {
		return err
	}

	r := img1.Bounds()
	s := r.Size()

	if s.X > 500 {
		delta = float32(s.X) / 500.0
	} else {
		delta = 1.0
	}

	nx := float32(s.X) / delta
	ny := float32(s.Y) / delta

	img2 := Resize(img1, r, int(nx), int(ny))

	fimg2, _ := os.Create(blog.Dir + "img/thumbs/" + file)
	jpeg.Encode(fimg2, img2, &jpeg.Options{jpeg.DefaultQuality})

	return nil
}
