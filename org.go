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
	"strings"
	"regexp"
	"strconv"
)





/*
 HTML conversion
*/

var head1Reg = regexp.MustCompile("(?m)^\\* (?P<head>.+)\\n")
var head2Reg = regexp.MustCompile("(?m)^\\*\\* (?P<head>.+)\\n")
var linkReg = regexp.MustCompile("\\[\\[(?P<url>[^\\]]+)\\]\\[(?P<text>[^\\]]+)\\]\\]")
var imgLinkReg = regexp.MustCompile("\\[\\[file:\\.\\./img/(?P<img>[^\\]]+)\\]\\[file:\\.\\./img/(?P<thumb>[^\\]]+)\\]\\]")
var imgReg = regexp.MustCompile("\\[\\[\\.\\./img/(?P<src>[^\\]]+)\\]\\]")

var codeReg = regexp.MustCompile("(?m)^\\#\\+BEGIN_SRC \\w*\\n(?P<code>(?s)[^\\#]+)^\\#\\+END_SRC\\n")
var codeHeaderReg = regexp.MustCompile("(?m)^\\#\\+BEGIN_SRC \\w*\\n")
var codeFooterReg = regexp.MustCompile("(?m)^\\#\\+END_SRC\\n")

var quoteReg = regexp.MustCompile("(?m)^\\#\\+BEGIN_QUOTE\\s*\\n(?P<cite>(?s)[^\\#]+)^\\#\\+END_QUOTE\\n")
var quoteHeaderReg = regexp.MustCompile("(?m)^\\#\\+BEGIN_QUOTE\\s*\\n")
var quoteFooterReg = regexp.MustCompile("(?m)^\\#\\+END_QUOTE\\n")

var centerReg = regexp.MustCompile("(?m)^\\#\\+BEGIN_CENTER\\s*\\n(?P<cite>(?s)[^\\#]+)^\\#\\+END_CENTER\\n")
var centerHeaderReg = regexp.MustCompile("(?m)^\\#\\+BEGIN_CENTER\\s*\\n")
var centerFooterReg = regexp.MustCompile("(?m)^\\#\\+END_CENTER\\n")

var parReg = regexp.MustCompile("\\n\\n+(?P<text>[^\\n]+)")
var allPropsReg = regexp.MustCompile(":PROPERTIES:(?s).+:END:")
var rawHTML = regexp.MustCompile("\\<A-Za-z[^\\>]+\\>")

//estilos de texto
var boldReg = regexp.MustCompile("(?P<prefix>[\\s|\\W]+)\\*(?P<text>[^\\s][^\\*]+)\\*(?P<suffix>[\\s|\\W]*)")
var italicReg = regexp.MustCompile("(?P<prefix>[\\s])/(?P<text>[^\\s][^/]+)/(?P<suffix>[^A-Za-z0-9]*)")
var ulineReg = regexp.MustCompile("(?P<prefix>[\\s|\\W]+)_(?P<text>[^\\s][^_]+)_(?P<suffix>[\\s|\\W]*)")
var codeLineReg = regexp.MustCompile("(?P<prefix>[\\s|\\W]+)=(?P<text>[^\\s][^\\=]+)=(?P<suffix>[\\s|\\W]*)")
var strikeReg = regexp.MustCompile("(?P<prefix>[\\s|[\\W]+)\\+(?P<text>[^\\s][^\\+]+)\\+(?P<suffix>[\\s|\\W]*)")


// listas
var ulistItemReg = regexp.MustCompile("\\s*[\\+|\\-]\\s+(?P<item>[^\\-|\\+|\\n+]+)")
var olistItemReg = regexp.MustCompile("\\s*[0-9]+\\.\\s+(?P<item>[^\\-|\\+]+)")
//var ulistItemReg = regexp.MustCompile("(?m)^\\s*[\\+|\\-]\\s+(?P<item>[^\\-|\\+]+)$")
//var olistItemReg = regexp.MustCompile("(?m)^\\s*[0-9]+\\.\\s+(?P<item>[^\\-|\\+]+)$")


//var ulistReg = regexp.MustCompile("(?P<items>(\\<uli-begin\\>[^\\<]+\\<uli-end\\>)+)")
//var olistReg = regexp.MustCompile("(?P<items>(\\<oli-begin\\>[^\\<]+\\<oli-end\\>)+)")
var ulistReg = regexp.MustCompile("(?P<items>(\\<uli-begin\\>.+\\<uli-end\\>)+)")
var olistReg = regexp.MustCompile("(?P<items>(\\<oli-begin\\>.+\\<oli-end\\>)+)")




func Org2HTML(content []byte,url string)(string){

	// First remove all HTML raw tags for security
	out:=rawHTML.ReplaceAll(content,[]byte(""))

	// headings (h1 is not admit in the post body)
	out=head1Reg.ReplaceAll(out,[]byte(""))
	out=head2Reg.ReplaceAll(out,[]byte("<h2>$head</h2>\n"))

	// images
	out=imgReg.ReplaceAll(out,[]byte("<a href='"+url+"/img/$src'><img src='"+url+"/img/thumbs/$src'/></a>"))
	out=imgLinkReg.ReplaceAll(out,[]byte("<a href='"+url+"/img/$img'><img src='"+url+"/img/thumbs/$thumb'/></a>"))
	out=linkReg.ReplaceAll(out,[]byte("<a href='$url'>$text</a>"))

	// Extract blocks codes
	codeBlocks,out:=extractBlocks(string(out),
		codeReg,
		codeHeaderReg,
		codeFooterReg,
		"code")

	quoteBlocks,out:=extractBlocks(string(out),
		quoteReg,
		quoteHeaderReg,
		quoteFooterReg,
		"quote")

	out=centerReg.ReplaceAll(out,[]byte("<span class=\"center\">$cite</span>\n"))
	out=parReg.ReplaceAll(out,[]byte("\n\n<p/>$text"))
	out=allPropsReg.ReplaceAll(out,[]byte("\n"))

	// font styles
	out=italicReg.ReplaceAll(out,[]byte("$prefix<i>$text</i>$suffix"))
	out=boldReg.ReplaceAll(out,[]byte("$prefix<b>$text</b>$suffix"))
	out=ulineReg.ReplaceAll(out,[]byte("$prefix<u>$text</u>$suffix"))
	out=codeLineReg.ReplaceAll(out,[]byte("$prefix<code>$text</code>$suffix"))
	out=strikeReg.ReplaceAll(out,[]byte("$prefix<s>$text</s>$suffix"))

	// List with fake tags for items
	out=ulistItemReg.ReplaceAll(out,[]byte("<uli-begin>$item<uli-end>"))
	out=ulistReg.ReplaceAll(out,[]byte("<ul>\n$items</ul>\n"))
	out=olistItemReg.ReplaceAll(out,[]byte("<fake-oli>$item</fake-oli>\n"))
	out=olistReg.ReplaceAll(out,[]byte("<ol>\n$items</ol>\n"))

	// Removing fake items tags
	sout:=string(out)
	sout=strings.Replace(sout,"<uli-begin>","<li>",-1)
	sout=strings.Replace(sout,"<uli-end>","</li>\n",-1)
	sout=strings.Replace(sout,"<oli-begin>","<li>",-1)
	sout=strings.Replace(sout,"<oli-end>","</li>\n",-1)

	// Reinsert block codes
	sout=insertBlocks(sout,codeBlocks,"<pre><code>","</code></pre>","code")
	sout=insertBlocks(sout,quoteBlocks,"<blockquote>","</blockquote>","quote")

	return sout
}



func extractBlocks(src string,fullReg,headerReg,footerReg *regexp.Regexp,name string)([][]byte,[]byte){
	out:=[]byte(src)
	blocks:=fullReg.FindAll(out,-1)
	for i:=range blocks{
		bstring:=string(blocks[i])
		blocks[i]=headerReg.ReplaceAll(blocks[i],[]byte("\n"))
		blocks[i]=footerReg.ReplaceAll(blocks[i],[]byte("\n"))
		out=[]byte(strings.Replace(string(out),bstring,"block"+name+":"+strconv.Itoa(i)+"\n",1))
	}

	return blocks,out
}


func insertBlocks(src string,blocks [][]byte,header,footer,name string)(string){
	s:=src
	for i:=range blocks{
		bstring:=string(blocks[i])
		s=strings.Replace(s,"block"+name+":"+strconv.Itoa(i)+"\n",
			header+bstring+footer+"\n",1)
	}

	return s
}