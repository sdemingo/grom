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
	"flag"
	"strings"
	"os"
)


/*
 grom commands
 */

const LICENSE = `Copyright (C) 2013  Sergio de Mingo 
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
`

const HELP =`    usage: grom [cmd] [args]
	      
	      - create     : Create a new blog
	      - build      : Build html files from the sources
              - clean      : Remove html files
              - serve      : Serve the blog on a builtin web service
	      - add-post   : Create a new post
	      - add-static : Create a new static page 
	      - help       : Show this message
`

func create_post(args []string){

	if (len(args)<2){
		fmt.Printf("grom add-post <post-id>\n")
		return
	}

	pwd,_:=os.Getwd()
	dir:=checkDirPath(pwd)
	title:=args[2]

	blog:=LoadBlog(dir)
	if (blog==nil){
		fmt.Printf("Error during blog load\n")
	}
	fmt.Printf("Load info from: %s\n",blog.Info["Name"])

	err:=blog.AddArticle(title)
	if err!=nil{
		fmt.Printf("Post not created: %s\n",err.Error())
	}else{
		fmt.Printf("Post created\n")
	}
}


func create_static(args []string){

	if (len(args)<2){
		fmt.Printf("grom add-static <page-id>\n")
		return
	}

	pwd,_:=os.Getwd()
	dir:=checkDirPath(pwd)
	title:=args[2]

	blog:=LoadBlog(dir)
	if (blog==nil){
		fmt.Printf("Error during blog load\n")
	}
	fmt.Printf("Load info from: %s\n",blog.Info["Name"])

	err:=blog.AddStaticPage(title)
	if err!=nil{
		fmt.Printf("Page not created: %s\n",err.Error())
	}else{
		fmt.Printf("Page created\n")
	}
}


func create_blog(args []string){

	if (len(args)<2){
		fmt.Printf("grom create <grom-dir>\n")
		return
	}

	tdir:=checkDirPath(args[1])
	pwd,_:=os.Getwd()
	bdir:=checkDirPath(pwd)
	blog,err:=CreateBlog(bdir,tdir)
	if (blog==nil){
		fmt.Printf("Error during blog creation: %s\n",err.Error())
	}
	
}

func serve_blog(args []string){

	pwd,_:=os.Getwd()
	dir:=checkDirPath(pwd)
	blog:=LoadBlog(dir)
	if (blog==nil){
		fmt.Printf("Error during blog load\n")
	}

	fmt.Printf("Load info from: %s\n",blog.Info["Name"])
	
	blog.Info["Url"]="http://localhost:9999"

	err:=blog.Build()
	if (err!=nil){
		fmt.Println(err)
	}else{
		fmt.Printf("Build blog succesfully\n")
	}

	warning:=`
        Blog has been build using a testing root URL. 
        Remember build again before push it on production.
`
	fmt.Printf(warning)
	fmt.Printf("\nServe blog on: %s\n",blog.Info["Url"])
	err=blog.Serve()
}


func build_blog(args []string){

	pwd,err:=os.Getwd()
	if err!=nil{
		fmt.Printf("Current directory is not Grom blog\n")
		return
	}
	dir:=checkDirPath(pwd)
	blog:=LoadBlog(dir)
	if (blog==nil){
		fmt.Printf("Error during blog load\n")
		return
	}

	fmt.Printf("Load info from: %s\n",blog.Info["Name"])
	
	err=blog.Build()
	if (err!=nil){
		fmt.Println(err)
	}else{
		fmt.Printf("Build blog succesfully\n")
	}

}

func clean_blog(args []string){

	pwd,err:=os.Getwd()
	if err!=nil{
		fmt.Printf("Current directory is not Grom blog\n")
		return
	}
	dir:=checkDirPath(pwd)
	if err=os.RemoveAll(dir+"tags");err!=nil{
		fmt.Println(err)
		return
	}
	os.Mkdir(dir+"/tags",0755)
	if err=os.RemoveAll(dir+"html");err!=nil{
		fmt.Println(err)
		return
	}
	os.Mkdir(dir+"/html",0755)
	if err=os.Remove(dir+"index.html");err!=nil{
		fmt.Println(err)
		return
	}
	fmt.Printf("Clean blog succesfully\n")
}


func help(args []string){
	fmt.Printf ("%s\n",HELP)
}

func checkDirPath(dir string)(string){
	if (!strings.HasSuffix(dir,"/")){
		return dir+"/"
	}
	return dir
}



func main() {

	flag.Parse()

	args:=flag.Args()
	if (len(args)<1){
		help(args)
		return
	}


	cmd:=flag.Arg(0)
	switch (cmd){
	case "add-post":
		create_post(args)

	case "add-static":
		create_static(args)
		
	case "create":
		create_blog(args)

	case "build":
		build_blog(args)

	case "serve":
		serve_blog(args)

	case "clean":
		clean_blog(args)

	default:
		help(args)
	}
}