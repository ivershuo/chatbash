package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

func init() {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		fmt.Println("此工具只能在macOS和Linux上运行")
		os.Exit(1)
	}
}

func main() {
	key := os.Getenv("OPENAI_KEY")
	if key == "" {
		fmt.Println("没有配置key，可在终端中使用以下方式配置：\nexport OPENAI_KEY=sk-xxxxxx")
		return
	}

	q := strings.Join(os.Args[1:], "")
	if q == "" {
		fmt.Println("您没有输入任何指令")
		return
	}

	openchat := NewChat(key, initMessage)

	respText, err := openchat.Completion(q + "\n要实现上面的功能该使用怎样的bash命令？\n请注意：在回答中仅仅包含bash代码内容，回复格式：```bash\n#code here\n```，回复内容以“```”开头，“```”结尾。如果上面的功能无法使用明确的bash指令实现，请回复“IDONTKNOW”。")
	if err != nil {
		fmt.Println("发生了一点错误，" + err.Error())
		return
	}
	if respText == "IDONTKNOW" {
		fmt.Println("你在想啥呢！不要为难你的电脑了😓。")
		return
	}

	bash, err := ParseBash(respText)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	output, err := ExecBash(bash)
	var msg2Explain string
	if err == nil {
		msg2Explain = fmt.Sprintf("返回信息如下：%s\n我不是很懂计算机，请帮忙解释一下这个结果，以便我能听懂。如果可能，请使用《银河系漫游指南》中马文的说话语气回复我，但是你不要给我透露你使用的是马文的语气。", output)
	} else {
		msg2Explain = fmt.Sprintf("我运行 \"%s\" 命令之后获得了以下错误信息：%s\n请帮忙解释说明了什么？", bash, err.Error())
	}

	c := make(chan string)
	e := make(chan error)
	go openchat.CompletionStream(msg2Explain, c, e)

	loop := true
	for loop {
		select {
		case msg, ok := <-c:
			if !ok {
				loop = false
				break
			}
			io.WriteString(os.Stdout, msg)
		case err := <-e:
			fmt.Println("AI罢工了，只获得了以下的结果，希望你能看得懂\n" + output + "\n" + err.Error())
			loop = false
		}
	}

	openchat.GetConversation()

	os.Stdout.Close()
}
