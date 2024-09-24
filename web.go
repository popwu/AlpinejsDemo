package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// 自定义模板解析器
func parseTemplate(filename string) (string, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	// 正则表达式匹配 <include file="xxx.html"></include> 或 <include file="xxx.html" />
	re := regexp.MustCompile(`<include\s+file="([a-zA-Z0-9./_-]+)"\s*/?>`)
	matches := re.FindAllStringSubmatch(string(content), -1)

	for _, match := range matches {
		includeFile := match[1]
		includeContent, err := ioutil.ReadFile(filepath.Join("html", includeFile))
		if err != nil {
			return "", err
		}
		// 替换 <include file="xxx.html"></include> 或 <include file="xxx.html" /> 为实际内容
		content = []byte(strings.ReplaceAll(string(content), match[0], string(includeContent)))
	}

	return string(content), nil
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 获取请求路径并转换为文件路径
		path := r.URL.Path
		if path == "/" {
			path = "/index"
		}
		filePath := filepath.Join("html", path+".html")

		// 检查文件是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		// 解析模板文件
		parsedContent, err := parseTemplate(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 创建模板并渲染
		tmpl, err := template.New("webpage").Funcs(template.FuncMap{
			"include": func(name string) (template.HTML, error) {
				content, err := ioutil.ReadFile(filepath.Join("html", name))
				if err != nil {
					return "", err
				}
				return template.HTML(content), nil
			},
		}).Parse(parsedContent)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, nil)
	})

	// 监听 8000 端口并启动服务器
	http.ListenAndServe(":8000", nil)
}
