package ui

import (
	"fmt"
	"khromalabs/keeper/internal/config"
)

type Ui interface {
	Create(func (map[string]string)(int64,error))(error)
	Init(string, map[string]interface{}, []string)(error);
	Print(string,func(string,int)([]interface{},int,error))(error);
	UpdateOrDelete(string,bool,func(string,int)([]interface{},int,error),func(map[string]string,bool)error)(error);
}

var conf *config.Config

func init() {
	conf = config.Get()
}

func Init(template string, templateData map[string]interface{}, templateKeys []string) (Ui,error) {
	var ui Ui
	switch conf.Ui {
	case "cli":
		ui = &UiCli{}
	default:
		return nil, fmt.Errorf("Unknown ui: %v", template)
	}
	return ui, ui.Init(template,templateData,templateKeys)
}
