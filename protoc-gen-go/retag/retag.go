package retag

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/generator"
)

func init() {
	generator.RegisterPlugin(new(retag))
}

type retag struct {
	gen         *generator.Generator
	tags        map[string]string
	fieldMaxLen int
	tagMaxLen   int
}

// Name returns the name of this plugin, "settag"
func (r *retag) Name() string {
	return "retag"
}

// Init initializes the plugin.
func (r *retag) Init(gen *generator.Generator) {
	r.gen = gen
}

func (r *retag) P(args ...interface{}) { r.gen.P(args...) }

// Generate generates code for the services in the given file.
func (r *retag) Generate(file *generator.FileDescriptor) {
	r.getStructTags(*file.Name)
	r.retag()
}

// GenerateImports generates the import declaration for this file.
func (r *retag) GenerateImports(file *generator.FileDescriptor) {}

func (r *retag) getStructTags(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	r.tags = make(map[string]string)
	var begin bool
	var comment bool
	var msgName string
	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}

		if strings.HasPrefix(strings.TrimSpace(string(line)), "/*") {
			comment = true
		}

		if comment && strings.Contains(string(line), "*/") {
			comment = false
			continue
		}

		if comment {
			continue
		}

		if strings.HasPrefix(strings.TrimSpace(string(line)), "message") {
			begin = true
			msgName = strings.Fields(string(line))[1]
			continue
		}

		if begin == true && line[0] == '}' {
			begin = false
			continue
		}

		if begin == true {
			if strings.HasPrefix(strings.TrimSpace(string(line)), "//") {
				continue
			}

			k, v := getFieldTag(string(line), msgName)
			r.tags[k] = v

			if len(strings.Split(k, ".")[1]) > r.fieldMaxLen {
				r.fieldMaxLen = len(strings.Split(k, ".")[1])
			}

			tags := strings.Fields(v)
			for _, tag := range tags {
				if len(strings.Split(tag, ":")[1])-2 > r.tagMaxLen {
					r.tagMaxLen = len(strings.Split(tag, ":")[1]) - 2
				}
			}
		}
	}
}

func getFieldTag(line string, msgName string) (field string, tag string) {
	fts := strings.Split(line, "//")
	tag = fts[1]
	fs := strings.Fields(fts[0])
	fsl := len(fs)
	field = msgName + "."
	for i := 0; i < fsl; i++ {
		if i == fsl-1 {
			field += fs[i]
			break
		} else {
			if fs[i+1] == "=" {
				field += fs[i]
				break
			}
		}
	}

	tag = strings.TrimSpace(tag)
	tag = strings.Trim(tag, "`")
	tag = trimInside(tag)

	return
}

func trimInside(s string) string {
	for {
		if strings.Contains(s, "  ") {
			s = strings.Replace(s, "  ", " ", -1)
		} else {
			break
		}
	}

	return s
}

func (r *retag) retag() {
	if len(r.tags) <= 0 {
		return
	}

	readbuf := bytes.NewBuffer([]byte{})
	readbuf.Write(r.gen.Buffer.Bytes())
	buf := bytes.NewBuffer([]byte{})

	reader := bufio.NewReader(readbuf)
	var begin bool
	var comment bool
	var msgName string
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			buf.WriteString("\n")
			break
		}

		if strings.HasPrefix(strings.TrimSpace(string(line)), "/*") {
			comment = true
		}

		if comment && strings.Contains(string(line), "*/") {
			comment = false
			buf.Write(line)
			buf.WriteString("\n")
			continue
		}

		if comment {
			buf.Write(line)
			buf.WriteString("\n")
			continue
		}

		if r.needRetag(strings.TrimSpace(string(line))) {
			begin = true
			msgName = strings.Fields(string(line))[1]
			buf.Write(line)
			buf.WriteString("\n")
			continue
		}

		if begin == true && line[0] == '}' {
			begin = false
			buf.Write(line)
			buf.WriteString("\n")
			continue
		}

		if begin == true {
			if strings.HasPrefix(strings.TrimSpace(string(line)), "//") {
				buf.Write(line)
				buf.WriteString("\n")
				continue
			}

			fields := strings.Fields(strings.TrimSpace(string(line)))
			key := msgName + "." + fields[0]
			tag := r.tags[key]
			newline := resetTag(string(line), fields[0], tag, r.fieldMaxLen, r.tagMaxLen)
			buf.WriteString(newline)
			buf.WriteString("\n")
			continue
		}
		buf.Write(line)
		buf.WriteString("\n")
	}

	r.gen.Buffer.Reset()
	data := buf.Bytes()
	r.gen.Buffer.Write(data)
}

func (r *retag) needRetag(line string) bool {
	for k := range r.tags {
		ks := strings.Split(k, ".")
		sub := "type " + ks[0] + " struct"
		if strings.HasPrefix(line, sub) {
			return true
		}
	}

	return false
}
func resetTag(line string, field string, tag string, maxlenField, maxlenTag int) string {
	//reset default json
	res := strings.Trim(strings.TrimRight(strings.TrimRight(line, "\n"), " "), "`")
	if strings.Contains(line, "json:") && strings.Contains(tag, "json:") {
		substr := " json:\"" + field + ",omitempty\""
		res = strings.Replace(res, substr, "", -1)
	}

	fs := strings.Fields(res)
	for i := 2; i < len(fs); i++ {
		if i == 2 {
			res = strings.Replace(res, fs[i], "`", -1)
			fs[i] = fs[i][1:]
		} else {
			res = strings.Replace(res, fs[i], "", -1)
		}
	}

	fs = append(fs, strings.Fields(tag)...)

	for i := 2; i < len(fs); i++ {
		if i == 2 {
			format := "%-" + strconv.Itoa(len(`protobuf:"bytes,xxx,opt,name=`)+maxlenField) + "s  "
			res += fmt.Sprintf(format, fs[i])
		} else if i != len(fs)-1 {
			format := "%-" + strconv.Itoa(len(fs[i])-len(strings.Trim(strings.Split(fs[i], ":")[1], "\""))+maxlenTag) + "s  "
			res += fmt.Sprintf(format, fs[i])
		} else {
			res += fs[i]
		}
	}

	res += "`"

	return res
}
