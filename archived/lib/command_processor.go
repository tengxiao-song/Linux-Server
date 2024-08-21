package command_processor

import (
	"bytes"
	"os/exec"
	"strings"
)

type CP struct {
}

func (cp CP) RunCmd(cmd string) (string, error) {

	cmdObj := exec.Command("sh", "-c", cmd) // 调用api创建命令对象
	var outBuf bytes.Buffer
	cmdObj.Stdout = &outBuf                        // 将io输入重定向到缓冲区
	err := cmdObj.Run()                            // 执行方法并得到error
	return strings.TrimSpace(outBuf.String()), err // 返回输出和error
}

// 静态方法
func RunCmd(cmd string) (string, error) {
	// time.Sleep(5 * time.Second)
	cp := CP{}
	return cp.RunCmd(strings.Split(cmd, " ")[1])
}
