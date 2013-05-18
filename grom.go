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
	      
	      - create-post   : Create a new post
	      - create-static : Create a new static page 
	      - create-blog   : Create a new blog
	      - build-blog    : Build html files from the sources
   	      - clean-blog    : Remove all html files
	      - help          : Show this message
`

func create_post(args []string){
	var t string
	if (len(args)<3){
		fmt.Printf("grom create-post <blog-dir> <post-id>\n",t)
		return
	}

	dir:=checkDirPath(args[1])
	title:=args[2]

	blog:=LoadBlog(dir)
	if (blog==nil){
		fmt.Printf("Error during blog load\n")
	}
	fmt.Printf("Load info from: %s\n",blog.Info["Name"])

	err:=blog.AddArticle(title)
	if err!=nil{
		fmt.Printf("Post not created: %s\n",err)
	}else{
		fmt.Printf("Post created\n")
	}
}


func create_static(args []string){
	var t string
	if (len(args)<3){
		fmt.Printf("grom create-static <blog-dir> <page-id>\n",t)
		return
	}

	dir:=checkDirPath(args[1])
	title:=args[2]

	blog:=LoadBlog(dir)
	if (blog==nil){
		fmt.Printf("Error during blog load\n")
	}
	fmt.Printf("Load info from: %s\n",blog.Info["Name"])

	err:=blog.AddStaticPage(title)
	if err!=nil{
		fmt.Printf("Page not created: %s\n",err)
	}else{
		fmt.Printf("Page created\n")
	}
}


func create_blog(args []string){

	if (len(args)<3){
		fmt.Printf("grom create-blog <grom-dir> <new-blog-dir>\n")
		return
	}

	tdir:=checkDirPath(args[1])
	bdir:=checkDirPath(args[2])
	blog,err:=CreateBlog(bdir,tdir)
	if (blog==nil){
		fmt.Printf("Error during blog creation: %s\n",err.Error())
	}
	
}

func clean_blog(args []string){
	var t string
	if (len(args)<2){
		fmt.Printf("grom clean-blog <blog-dir>\n",t)
		return
	}
	dir:=checkDirPath(args[1])
	blog:=LoadBlog(dir)
	if (blog==nil){
		fmt.Printf("Error during blog load\n")
	}

	fmt.Printf("Load info from: %s\n",blog.Info["Name"])
	
	err:=blog.Clean()
	if (err!=nil){
		fmt.Println(err)
	}else{
		fmt.Printf("Clean blog succesfully\n")
	}
}


func build_blog(args []string){
	var t string
	if (len(args)<2){
		fmt.Printf("grom build-blog <blog-dir>\n",t)
		return
	}
	dir:=checkDirPath(args[1])
	blog:=LoadBlog(dir)
	if (blog==nil){
		fmt.Printf("Error during blog load\n")
		return
	}

	fmt.Printf("Load info from: %s\n",blog.Info["Name"])
	
	err:=blog.Build()
	if (err!=nil){
		fmt.Println(err)
	}else{
		fmt.Printf("Build blog succesfully\n")
	}

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
	case "create-post":
		create_post(args)

	case "create-static":
		create_static(args)
		
	case "create-blog":
		create_blog(args)

	case "build-blog":
		build_blog(args)

	default:
		help(args)
	}
}