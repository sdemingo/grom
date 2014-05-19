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
	ArticleTags Tags
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
	a.Meta["Tags"]=""

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
	a.Meta["Tags"]=parseProperty(a.Content,"Tags")
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

	a.ArticleTags=make(Tags)
	
	return a,nil
}



func (a *Article)GetValidId()(string){
	if a==nil {
		return ""
	}
	d:=a.GetDate()
	ds:=d.Format("01")
	return ds+"-"+a.Id
}

func (a *Article)GetYear()(string){
	if a==nil {
		return ""
	}
	d:=a.GetDate()
	ds:=d.Format("2006")
	return ds
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

