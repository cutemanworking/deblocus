// -----------------  update-webpanel.go  -----------------
// The directory `static` is just for keeping static resources and tools.
// This file will NOT be compiled into deblocus executable file.
// This is a independent tool for generating or updating the `deblocus/tunnel/webpanel.go`.
// Usage:
//     cd deblocus/static/update-webpanel
//     go run update-webpanel.go
// -----------------------------------------------------
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const (
	src_file  = "webpanel.html"
	out_file  = "../../tunnel/webpanel.go"
	self_name = "update-webpanel.go"
	self_dir  = "deblocus/static/update-webpanel"
)

func IsNotExist(file string) bool {
	_, err := os.Stat(file)
	return os.IsNotExist(err)
}

func main() {
	if IsNotExist(self_name) {
		log.Fatalln("Please change cwd to", self_dir)
	}

	mainPage, err := ioutil.ReadFile(src_file)
	if err != nil {
		log.Fatalln(err)
	}
	out, err := os.OpenFile(out_file, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	defer out.Close()

	fmt.Fprint(out, pkg_line, "\n\n")
	// 404
	fmt.Fprint(out, "const _TPL_HTTP404_BODY = `", tpl_web404_body, "`\n")
	// mainpage
	fmt.Fprint(out, "const _TPL_MAIN_PAGE = `")
	out.Write(mainPage)
	fmt.Fprintln(out, "`")

	fmt.Println("Update done.")
}

const pkg_line = `package tunnel`

const tpl_web404_body = `<html>
<head><title>404 Not Found</title></head>
<body>
<center><h1>404 Not Found</h1></center>
<hr><center>%s</center>
</body>
</html>
`
