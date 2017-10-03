package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

import (
	"github.com/bwmarrin/discordgo"
	"github.com/itokatsu/nanogo/botutils"
)

var saveDir = "./.saves"

type Plugin interface {
	HandleMsg(cmd *botutils.Cmd, s *discordgo.Session, m *discordgo.MessageCreate)
	Help() string
	Name() string
}

type HasSave interface {
	Load([]byte) error
	Save() []byte
}

type HasCleanup interface {
	Cleanup()
}

// plugin Handler
type pluginHandler struct {
	plugins map[string]Plugin
}

func (ph *pluginHandler) Load(p Plugin) Plugin {
	psl, ok := p.(HasSave)
	if !ok {
		return ph.Start(p)
	}
	filename := fmt.Sprintf("%v/%v", saveDir, p.Name())
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return ph.Start(p)
	}
	if err := psl.Load(content); err != nil {
		return ph.Start(p)
	}
	fmt.Printf("+ Plugin '%v' loaded \n", p.Name())
	return ph.Start(p)
}

func (ph *pluginHandler) Start(p Plugin) Plugin {
	ph.plugins[p.Name()] = p
	fmt.Printf("+ Plugin '%v' started \n", p.Name())
	return p
}

func (ph *pluginHandler) Save() {
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		err := os.Mkdir(saveDir, 0777)
		if err != nil {
			fmt.Println(err)
		}
	}
	for _, p := range ph.plugins {
		psl, ok := p.(HasSave)
		if !ok {
			continue
		}
		filename := fmt.Sprintf("%v/%v", saveDir, p.Name())
		data := psl.Save()
		if data == nil {
			continue
		}
		if err := ioutil.WriteFile(filename, data, 0777); err != nil {
			continue
		}
		fmt.Sprintf("* Plugin '%v' saved \n", p.Name())
	}
	ph.Cleanup()
}

func (ph *pluginHandler) Cleanup() {
	for _, p := range ph.plugins {
		pclnup, ok := p.(HasCleanup)
		if !ok {
			continue
		}
		pclnup.Cleanup()
	}
}
