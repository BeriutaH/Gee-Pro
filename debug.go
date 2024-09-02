package GeeRPC

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

const debugText = `<html>
	<body>
	<title>GeeRPC Services</title>
	{{range .}}
	<hr>
	Service {{.Name}}
	<hr>
		<table>
		<th align=center>Method</th><th align=center>Calls</th>
		{{range $name, $mtype := .Method}}
			<tr>
			<td align=left font=fixed>{{$name}}({{$mtype.ArgType}}, {{$mtype.ReplyType}}) error</td>
			<td align=center>{{$mtype.NumCalls}}</td>
			</tr>
		{{end}}
		</table>
	{{end}}
	</body>
	</html>`

// 创建一个模板名字是 "RPC debug" 新的模板, Parse 方法用于解析模板定义
var debug = template.Must(template.New("RPC debug").Parse(debugText))

type debugHTTP struct {
	*Server
}

type debugService struct {
	Name   string
	Method map[string]*methodType
}

func (d debugHTTP) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var services []debugService
	d.serviceMap.Range(func(namei, svci any) bool {
		svc := svci.(*service)
		services = append(services, debugService{
			Name:   namei.(string),
			Method: svc.method,
		})
		return true
	})

	mapserve := services[0].Method
	for k := range mapserve {
		log.Printf("方法对象具体信息: %+v\n", mapserve[k])
	}

	err := debug.Execute(w, services)
	if err != nil {
		// Fprintln 用于将格式化的字符串输出到指定的 io.Writer 接口，并在末尾添加换行符
		_, _ = fmt.Fprintln(w, "rpc: 执行模板时出错: ", err.Error())
	}
}
