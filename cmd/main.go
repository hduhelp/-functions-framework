package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"plugin"
	"syscall"
	"time"

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
	log.Println("building plugin so")
	cmd := exec.Command("go",
		"build", "-buildmode=plugin",
		"-o", soPath(pluginPath),
		soDir(pluginPath),
	)
	log.Println(cmd)
	p, err := cmd.Output()
	log.Println(string(p))
	return err
}

func LoadAndInvokeSomethingFromPlugin(pluginPath string) (f functions.GroupHandler, err error) {
	var p *plugin.Plugin

	files, err := os.ReadDir(pluginPath)
	var soUpdateAt time.Time
	var codeUpdateAt time.Time
	for _, v := range files {
		info, _ := v.Info()
		stat := info.Sys().(*syscall.Stat_t)
		switch {
		case filepath.Ext(v.Name()) == ".so":
			soUpdateAt = time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
		case filepath.Ext(v.Name()) == ".go":
			if t := time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec)); t.After(codeUpdateAt) {
				codeUpdateAt = t
			}
		}
	}
	if codeUpdateAt.After(soUpdateAt) {
		if err := BuildPlugin(pluginPath); err != nil {
			return nil, err
		}
	}
	p, err = plugin.Open(soPath(pluginPath))
	for err != nil {
		if err := BuildPlugin(pluginPath); err != nil {
			return nil, err
		} else {
			p, err = plugin.Open(soPath(pluginPath))
		}
	}
	f1, err := p.Lookup("Instance")
	if err != nil {
		return nil, err
	}
	i, ok := f1.(functions.GroupHandler)
	if !ok {
		return nil, fmt.Errorf("%s.Instance does not implement Function", path.Dir(pluginPath))
	}

	return i, nil
}
