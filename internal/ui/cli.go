package ui

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/peterh/liner"
	"golang.org/x/crypto/ssh/terminal"
	. "khromalabs/keeper/internal/log"
)
type UiCli struct {
	Template string;
	TemplateSchema map[string]interface{};
	TemplateKeys []string;
}

func (u *UiCli) Init(template string, templateSchema map[string]interface{}, templateKeys []string) error {
	u.Template = template
	u.TemplateSchema = templateSchema
	u.TemplateKeys = templateKeys
	return nil
}

func (u *UiCli) input(init map[string]string) (map[string]string,error) {
	var err error
	out := make(map[string]string, 0)
	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	reader := bufio.NewReader(os.Stdin)
	if init != nil {
		out = init
	}
	for {
		for _, label := range u.TemplateKeys {
			attr := u.TemplateSchema[label].(map[string]interface{})
			var validation map[string]interface{}
			required := false
			regex := ""
			tip := ""
			if _, has := attr["validation"]; has {
				validation = attr["validation"].(map[string]interface{})
				if _, has = validation["required"]; has {
					required = validation["required"].(bool)
				}
				if _, has = validation["regex"]; has {
					regex = validation["regex"].(string)
				}
				if _, has = validation["tip"]; has {
					tip = validation["tip"].(string)
				}
			}
			var prefix string
			if required {
				prefix = ">*%s: "
			} else {
				prefix = "> %s: "
			}
			for {
				if attr["type"] == "autodate" {
					out[label] = time.Now().Format(time.DateOnly)
					fmt.Printf("  %s: %s\n", strings.Title(label), out[label])
				} else if attr["type"] == "text" {
					prompt := fmt.Sprintf(prefix,strings.Title(label))
					fmt.Printf("%s[Open external editor (Y/n)] ",prompt)
					c,_,_ := reader.ReadRune()
					fmt.Print(string(c) + "\n")
					if c != 'n' && c != 'N' {
						out[label],err = u.getEditorOutput(out[label])
						if err != nil {
							return nil,fmt.Errorf("Can't open the text editor: %s", err)
						}
					}
				} else {
					prompt := fmt.Sprintf(prefix,strings.Title(label))
					out[label],err = line.PromptWithSuggestion(prompt, out[label], -1)
					if attr["type"] == "tokens" && out[label] != "" {
						regex = "^([ ]*[A-z0-9]+[ ]*[,]?)+$"
						tip = "tag1[,tag2,tag3]"
					}
				}
				if err == liner.ErrPromptAborted {
					return nil,fmt.Errorf("Aborted")
				}
				cancontinue := true
				if required && out[label] == "" {
					fmt.Println("!Required value for " + label)
					cancontinue = false
				} else if regex != "" {
					match, err := regexp.MatchString(regex, out[label]);
					if err != nil || !match {
						fmt.Println("!Invalid format for " + label)
						if tip != "" {
							fmt.Println("!Expected format: " + tip)
						}
						cancontinue = false
					}
				}
				if cancontinue {
					break
				}
			}
		}
		fmt.Print("> Ready? (y/N) ")
		c,_,_ := reader.ReadRune()
		fmt.Print(string(c) + "\n")
		if c == 'y' || c == 'Y' {
			break
		}
	}
	return out,nil
}

func (u *UiCli) getEditorOutput(input string) (string,error) {
	tmpDir := os.TempDir()
	tmpFile, tmpFileErr := ioutil.TempFile(tmpDir, "keeper")
	defer os.Remove(tmpFile.Name())
	if tmpFileErr != nil {
		return "",fmt.Errorf("Error %s while creating tempFile", tmpFileErr)
	}
	path, err := exec.LookPath(conf.Editor)
	if err != nil {
		return "",fmt.Errorf("Error %s while looking up for %s", err, conf.Editor)
	}
	if input != "" {
		_, err := tmpFile.Write([]byte(input))
		if err != nil {
			return "",fmt.Errorf("Start failed: %s", err)
		}
	}
	cmd := exec.Command(path, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return "",fmt.Errorf("Start failed: %s", err)
	}
	err = cmd.Wait()
	if err != nil {
		return "",fmt.Errorf("Command finished with error: %v\n", err)
	}
	v, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "",fmt.Errorf("Error can't read file %s: %s", path, err)
	}
	return string(v),nil
}

func (u *UiCli) Print(filter string, read func(string, int)([]interface{},int,error)) (error) {
	output := fmt.Sprintf("Reading %s registries:\n\n", u.Template)
	var p []interface{}
	var err error
	var plen int
	i := 0
	if p, i, err = read(filter, i); err != nil {
		return err
	}
	if p == nil || len(p) == 0 {
		output += fmt.Sprintf("No registries found.\n")
		fmt.Print(output)
		return nil
	} else {
		if plen = len(p); plen == 1 {
			output += fmt.Sprintf("Found 1 registry\n")
		} else {
			output += fmt.Sprintf("Found %d registries\n", plen)
		}
	}
	for i := 0; i < plen; i += 1 {
		output += fmt.Sprintln("----")
		output += u.getRegistry(p[i].(map[string]string))
	}
	lines := strings.Count(output, "\n")
	_, terminalHeight, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return err
	}
	if lines >= terminalHeight {
		cmd := exec.Command(conf.Pager)
		cmd.Stdin = strings.NewReader(output)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return err
		}
	} else {
		fmt.Print(output)
	}
	return nil
}

func (u *UiCli) getRegistry(row map[string]string) (output string) {
	output += fmt.Sprintf("id: %s\n", row["id"])
	for _, label := range u.TemplateKeys {
		if field, ok := row[label]; ok  {
			schemafield := u.TemplateSchema[label].(map[string]interface{})
			if schemafield["type"] == "text" {
				output += fmt.Sprintln(label + ":\n" + field)
			} else {
				output += fmt.Sprintln(label + ":", field)
			}
		}
	}
	return output
}

func (u *UiCli) UpdateOrDelete(filter string, opUpdate bool, read func(string, int)([]interface{},int,error), opfunc func(map[string]string,bool)error)(error) {
	var err error
	var p []interface{}
	prefix := "\n> "
	var opstr string
	var actiontip string
	var plen int
	if opUpdate {
		fmt.Println("Updating registries for ", u.Template)
		opstr = "Update"
		actiontip = "(Y/n)"
	} else {
		fmt.Println("Deleting registries for ", u.Template)
		opstr = "Delete"
		actiontip = "(y/N)"
	}
	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	reader := bufio.NewReader(os.Stdin)
	i := 0
	if p, _, err = read(filter, i); err != nil {
		return err
	}
	if p == nil || len(p) == 0 {
		fmt.Printf("No registries found.")
		return nil
	} else {
		if plen = len(p); plen == 1 {
			fmt.Println("Found 1 registry:")
		} else {
			fmt.Println("Found", plen, "registries:")
		}
	}
	for i := 0; i < len(p); i += 1 {
		fmt.Println("\n----")
		row := p[i].(map[string]string)
		fmt.Print(u.getRegistry(row))
		fmt.Printf("%s%s registry? %s ", prefix, opstr, actiontip)
		c,_,_ := reader.ReadRune()
		if ( opUpdate && c != 'n' && c != 'N' ) || ( !opUpdate && c == 'y' || c == 'Y' ) {
			fmt.Println()
			if opUpdate {
				if _, err := u.input(row); err != nil {
					return fmt.Errorf("Input: %s", err)
				}
				LogD.Printf("%#v", row)
			}
			if err := opfunc(row,opUpdate); err != nil {
				return fmt.Errorf("%s: %s", opstr, err)
			}
			fmt.Printf("Registry %sd",strings.ToLower(opstr))
		} else {
			fmt.Printf("\nRegistry skipped")
		}
	}
	return nil
}

func (u *UiCli) Create(create func(map[string]string)(int64,error)) (error) {
	fmt.Printf("Creating new %s registry:\n", u.Template)
	registry, err := u.input(nil)
	if err != nil {
		return err
	}
	// LogD.Printf("fields: %+v\n", templateData)
	id, err := create(registry)
	if err != nil {
		return err
	}
	fmt.Printf("Created new %s registry with id %d\n", u.Template, id)
	return nil
}
