[![Go Report Card](https://goreportcard.com/badge/github.com/zupzup/calories)](https://goreportcard.com/report/github.com/zupzup/calories)

# blog-generator

A static blog generator that collects posts written in Markdown from the local file system (OR a configurable GitHub repository). A post has metadata in a YAML fie header. The code that this is forked from has [this](https://github.com/zupzup/blog) is an example repo for the blog at [https://zupzup.org/](https://zupzup.org/).

## Features

* Listing
* Sitemap Generator
* RSS Feed
* Code Highlighting
* Archive 
* Configurable Static Pages 
* Tags 
* File-Based Configuration

## Installation

```bash
go get github.com/fbaube/bloggenator
```

## Usage & Customization

### Configuration

The tool can be configured using a config file called `bloggen.yml`. There is a `bloggen.dist.yml.FS` in the repository you can use as a template.

Example Config File: (THIS EXAMPLE CONFIG IS OBSOLETE!)

```yml
generator:
    repo: 'https://github.com/zupzup/blog'
    tmp: 'tmp'
    dest: 'www'
blog:
    url: 'https://www.zupzup.org'
    language: 'en-us'
    description: 'A blog about Go, JavaScript, Open Source and Programming in General'
    dateformat: '02.01.2006'
    title: 'zupzup'
    author: 'Mario Zupan'
    frontpageposts: 10
    statics:
        files:
            - src: 'static/favicon.ico'
              dest: 'favicon.ico'
            - src: 'static/robots.txt'
              dest: 'robots.txt'
            - src: 'static/about.png'
              dest: 'about.png'
        templates:
            - src: 'static/about.html'
              dest: 'about'
```

### Running

Just execute

```bash
bloggenator
```

### Templates

Edit templates in `static` folder to your needs.

## Example Blog Repository

(OLD) [Blog](https://github.com/zupzup/blog)
