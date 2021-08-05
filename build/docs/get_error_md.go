package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

const (
	OK = iota
	InternalServerErr
	BadRequest
	NotFound
)

// getCodeSlice 获取错误数组
func getCodeSlice() (codeSlice []int) {
	var rootPath = "./internal/types/status"
	file, err := os.Open(rootPath)
	if err != nil {
		log.Panicf("open http dir err: %v", err)
	}

	defer file.Close()

	fileInfos, err := file.Readdir(-1)
	if err != nil {
		log.Panicf("readDir err:%v", err)
	}

	for _, fileInfo := range fileInfos {
		path := fmt.Sprintf("%s/%s", rootPath, fileInfo.Name())
		fileInfo, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			log.Panicf("open path err:%s", err)
		}

		var count int
		scanner := bufio.NewScanner(fileInfo)
		var strNum int

		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "iota") {
				strs := strings.Split(scanner.Text(), " ")
				strNum, _ = strconv.Atoi(strs[(len(strs) - 1)])
			}

			if strings.Contains(scanner.Text(), "errors.New") {
				count++
				code := strNum + count - 1
				codeSlice = append(codeSlice, code)
			}
		}
		fileInfo.Close()
	}
	return codeSlice
}

// writeCodeFile 将错误码输出到md文件中
func writeCodeFile() {
	errorSlice := getCodeSlice()
	file, err := os.Create("./error.md")
	if err != nil {
		log.Panicf("create file err:%v", err)
	}

	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintln(w, fmt.Sprintf("## 错误码列表\n### 通用"))
	fmt.Fprintf(w, "**%v: %v**  \n", OK, errors.GetCodeReason(OK))
	fmt.Fprintf(w, "**%v: %v**  \n", InternalServerErr, errors.GetCodeReason(InternalServerErr))
	fmt.Fprintf(w, "**%v: %v**  \n", BadRequest, errors.GetCodeReason(BadRequest))
	fmt.Fprintf(w, "**%v: %v**  \n", NotFound, errors.GetCodeReason(NotFound))

	for _, v := range errorSlice {
		switch v {
		case 1000:
			fmt.Fprintf(w, "### 家庭/物业\n**%v: %v**  \n", v, errors.GetCodeReason(v))
		case 2000:
			fmt.Fprintf(w, "### 设备\n**%v: %v**  \n", v, errors.GetCodeReason(v))
		case 3000:
			fmt.Fprintf(w, "### 房间/位置\n**%v: %v**  \n", v, errors.GetCodeReason(v))
		case 4000:
			fmt.Fprintf(w, "### 场景\n**%v: %v**  \n", v, errors.GetCodeReason(v))
		case 5000:
			fmt.Fprintf(w, "### 用户\n**%v: %v**  \n", v, errors.GetCodeReason(v))
		default:
			fmt.Fprintf(w, "**%v: %v**  \n", v, errors.GetCodeReason(v))
		}

		w.Flush()

	}
}

func main() {
	writeCodeFile()
}
