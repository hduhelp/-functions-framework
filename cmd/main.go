package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"plugin"

	"github.com/gin-gonic/gin"
	"github.com/hduhelp/functions-framework"
)

func main() {
	flag.Parse()
	pluginPath := flag.Arg(0)
	fc, err := LoadAndInvokeSomethingFromPlugin(pluginPath)
	if err != nil {
		log.Fatal(err)
	}
	router := gin.Default()
	fc.Handle(&router.RouterGroup)
	router.Run(":8880")
}

func soPath(pluginPath string) string {
	wd, _ := os.Getwd()
	return path.Join(
		wd,
		path.Base(pluginPath),
		fmt.Sprintf("%s.so", path.Dir(pluginPath)),
	)
}

func soDir(pluginPath string) string {
	dir, _ := path.Split(soPath(pluginPath))
	return dir
}

func BuildPlugin(pluginPath string) error {
	cmd := exec.Command("go",
		"build", "-buildmode=plugin",
		"-o", soPath(pluginPath),
		soDir(pluginPath),
	)
	p, err := cmd.CombinedOutput()
	log.Println(string(p))
	return err
}

func LoadAndInvokeSomethingFromPlugin(pluginPath string) (functions.Function, error) {
	p, err := plugin.Open(soPath(pluginPath))
	if err != nil {
		if err := BuildPlugin(pluginPath); err != nil {
			return nil, err
		}
	}

	f1, err := p.Lookup("Instance")
	if err != nil {
		return nil, err
	}
	i, ok := f1.(functions.Function)
	if !ok {
		return nil, fmt.Errorf("%s.Instance does not implement Function", path.Dir(pluginPath))
	}

	return i, nil
}
