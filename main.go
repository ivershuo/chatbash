package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"unicode"
)

var logPath string
var discardLog bool

func chineseAppend(text string) string {
	isChinese := false
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			isChinese = true
			break
		}
	}
	if isChinese {
		return "\n请使用中文回复我。"
	} else {
		return ""
	}
}

func init() {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		log.Fatal("This tool only works on macOS and Linux")
	}

	flag.StringVar(&logPath, "log", "", "log path, default is \"~/Library/Logs/chatbash\" for macOS, and \"~/.log/chatbash\" for Linux ")
	flag.BoolVar(&discardLog, "discard-log", false, "set to discard log")
}

func main() {
	key := os.Getenv("OPENAI_KEY")
	if key == "" {
		log.Fatal("No key was configured, it can be configured in the terminal or .bashrc by:\nexport OPENAI_KEY=sk-xxxxxx")
		return
	}

	flag.Parse()
	setLog(logPath, discardLog)

	q := strings.Join(os.Args[1:], "")
	if q == "" {
		fmt.Println("The command you entered is invalid.")
		return
	}
	InfoLog("Launch command：%s", q)

	openchat := NewChat(key, initMessage)

	respText, err := openchat.Completion(q + "\nwhat bash command should be used to achieve the above function? Please note: the answer should only contain bash code. Reply format: ```bash\n#code here\n```, and put the bash code in one line. The reply content starts with \"```\" and ends with \"```\". If the above function cannot be achieved using a specific bash command, please reply with \"[[IDONTKNOW]]\".")
	if err != nil {
		ErrLog("got error: %s", err)
		fmt.Println("An error has occurred: " + err.Error())
		return
	}
	InfoLog("got response text：%s", respText)
	if respText == "[[IDONTKNOW]]" {
		fmt.Println("What are you doing! Don't funning your computer like this 😓.")
		return
	}

	bash, err := ParseBash(respText)
	fmt.Println("bash", bash)
	if err != nil {
		ErrLog("parse bash error: %s", err)
		fmt.Println(err.Error())
		return
	}
	output, err := ExecBash(bash)
	var msg2Explain string
	if err == nil {
		InfoLog("run bash result:\n%s", output)
		msg2Explain = fmt.Sprintf("The output of the command is: %s\n. Please describe and explain the result using human-understandable language. (Please reply to me in the speaking style of Marvin from 'The Hitchhiker's Guide to the Galaxy,' but be sure not to reveal to me that you are using Marvin's tone.)"+chineseAppend(q), output)
	} else {
		ErrLog("run bash got error:\n%s", err)
		msg2Explain = fmt.Sprintf("After running the \"%s\" command, I received the following error message: %s. Can you please explain what it means?"+chineseAppend(q), bash, err.Error())
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
			fmt.Println("The AI went on strike and only obtained the following results. I hope you can understand this.\n" + output + "\n" + err.Error())
			loop = false
		}
	}

	fmt.Print("\n")
	os.Stdout.Close()
}
