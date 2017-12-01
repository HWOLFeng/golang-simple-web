package main

import (
	"bufio"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
)

func main() {
	serveWeb()
}

var themeName = getThemeName()
var staticPages = populateStaticPages()

func serveWeb() {
	// 取代 http 改用 github.com/gorilla/mux 当做handle
	gorillaRoute := mux.NewRouter()

	// 类似资源拦截器（链），是有包含关系的
	gorillaRoute.HandleFunc("/", serveContent)
	gorillaRoute.HandleFunc("/{page_alias}", serveContent)
	gorillaRoute.HandleFunc("/{page_alias}/contact", serveContact)          // 这是静态的路径
	gorillaRoute.HandleFunc("/{page_alias}/dy/{name}", serveContactDynamic) //动态路径参数

	// 配置静态资源拦截
	http.HandleFunc("/imgs/", serveResource)
	http.HandleFunc("/css/", serveResource)
	http.HandleFunc("/js/", serveResource)

	http.Handle("/", gorillaRoute)
	http.ListenAndServe(":8080", nil)
}
func serveHome(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World!")) // 注意改成 []byte
}

func serveContact(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello HWOLF! This contact page"))
}

func serveContactDynamic(w http.ResponseWriter, r *http.Request) {
	urlParams := mux.Vars(r) // 获取 map (key,value)
	name := urlParams["name"]
	w.Write([]byte("Hello " + name))
}

func serveContent(w http.ResponseWriter, req *http.Request) {
	urlParams := mux.Vars(req)
	page_alias := urlParams["page_alias"]
	if page_alias == "" {
		page_alias = "home"
	}
	// 根据路径找静态 html 文件
	staticPage := staticPages.Lookup(page_alias + ".html")

	// 404
	if staticPage == nil {
		staticPage = staticPages.Lookup("404.html")
		w.WriteHeader(404)
	}
	// 遍历
	staticPage.Execute(w, nil)
}

func getThemeName() string {
	return "bs4"
}

// 获取pages下的资源，并日志打印
func populateStaticPages() *template.Template {
	result := template.New("templates")
	templatePath := new([]string)
	basePath := "pages"
	// Open
	templateFolder, _ := os.Open(basePath)
	defer templateFolder.Close()
	templatePathRaw, _ := templateFolder.Readdir(-1)

	for _, pageInfo := range templatePathRaw {
		log.Println(pageInfo.Name())
		*templatePath = append(*templatePath, basePath+"/"+pageInfo.Name())
	}

	result.ParseFiles(*templatePath...)
	return result
}

// css,js 等资源获取
func serveResource(w http.ResponseWriter, req *http.Request) {
	path := "public/" + themeName + req.URL.Path
	var contentType string
	// 根据path 设定类型,只能用 = 不能用 :=
	// https://golangtc.com/t/57ad41e5b09ecc76e30000f6
	if strings.HasSuffix(path, ".css") {
		contentType = "text/css; charset=utf-8"
	} else if strings.HasSuffix(path, ".png") {
		contentType = "image/png; charset=utf-8"
	} else if strings.HasSuffix(path, ".jpg") {
		contentType = "image/jpg; charset=utf-8"
	} else if strings.HasSuffix(path, ".js") {
		contentType = "application/javascript; charset=utf-8"
	} else {
		contentType = "text/plain; charset=utf-8"
	}
	log.Println(path)
	// 捕获file 和err
	file, err := os.Open(path)
	if err == nil {
		defer file.Close()
		w.Header().Add("Content-Type", contentType)
		// 读取资源文件
		br := bufio.NewReader(file)
		// 写到网页中去
		br.WriteTo(w)
	} else {
		w.WriteHeader(404)
	}
}
