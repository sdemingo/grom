What is Grom?
=============

Grom is a small static web content generator based on org-mode
syntax. The Org-mode is a major mode for Emacs created by Carsten
Dominik. Grom only supports a subset of org-mode syntax. With Grom you
can create all contents for your website or your blog with this syntax
and then Grom exports your source files to HTML building the site. 
You can get a short review about org-mode syntax at http://orgmode.org/orgguide.pdf

Grom is written in Go and if you want to build the binary you need to
have the Go environment installed. Get it from http://golang.org


Quick start
===========

You can create your first blog site with Grom building the sample site
which is included in the distribution. Type these commands on your shell:

      $ cd /grom/directory
      $ cd sample-blog
      $ grom build
      
You can run a builtin web service on grom to test the sample blog:

      $ grom serve

Now, open your browse and write the next link in the address bar:

     http://localhost:9999




Create a new site
=================

You can create a new site typing the next command, the <grom-dir> is the directory where grom was installed.

    grom create <grom-dir>

Now you can add a new post or a new static page using: 
    
    grom add-post  <post-name>	   
    grom add-static <static-page-name>	

To edit your new post open your text editor and load the file 
from <site-dir>/posts or <site-dir>/static.

Finally, to build all the blog or test it before update the 
master version you can type:

    grom build
    grom serve



