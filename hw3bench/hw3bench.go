package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	jlexer "github.com/mailru/easyjson/jlexer"
)

type userStruct struct {
	Browsers []string
	Name     string
	Email    string
}

func easyjson8664aee4DecodeGithubComStepikGolangFirstPartHw3bench(in *jlexer.Lexer, out *userStruct) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 4)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Browsers = append(out.Browsers, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "name":
			out.Name = string(in.String())
		case "email":
			out.Email = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *userStruct) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson8664aee4DecodeGithubComStepikGolangFirstPartHw3bench(&r, v)
	return r.Error()
}

func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	seenbrowsers := make(map[string]bool, 150)
	user := &userStruct{}
	fmt.Fprintln(out, "found users:")
	scanner := bufio.NewScanner(file)
	for idx := 0; scanner.Scan(); idx++ {

		err := user.UnmarshalJSON(scanner.Bytes())

		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false

		for _, browser := range user.Browsers {
			isAndroidInner := strings.Contains(browser, "Android")
			isAndroid = isAndroid || isAndroidInner
			isMSIEInner := strings.Contains(browser, "MSIE")
			isMSIE = isMSIE || isMSIEInner

			if isAndroidInner || isMSIEInner {
				seenbrowsers[browser] = true
			}
		}

		if isAndroid && isMSIE {
			fmt.Fprintf(out, "[%d] %s <%s>\n", idx, user.Name, strings.Replace(user.Email, "@", " [at] ", -1))
		}

	}
	fmt.Fprintln(out, "\nTotal unique browsers", len(seenbrowsers))
}

func main() {

}
