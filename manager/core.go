/*
 * Copyright (c) 2022
 * Author: Xiangyu Liu
 * File: core.go
 * Date: 2022/10/15 下午9:09
 */

package manager

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gohouse/gorose/v2"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"os"
	"os/exec"
	"strings"
)

type HisManager struct {
	dbFile string
	engin  *gorose.Engin
	table  string
	index  string
}

func (hm *HisManager) db() gorose.IOrm {
	return hm.engin.NewOrm()
}

func (hm *HisManager) Init(dbFile, table, index string) error {
	if _, err := os.Stat(dbFile); err == nil {
		hm.dbFile = dbFile
		hm.engin, err = gorose.Open(&gorose.Config{
			Driver: "sqlite3",
			Dsn:    dbFile,
		})
		if err != nil {
			return err
		} else {
			hm.table = table
			hm.index = index
			return nil
		}
	} else {
		return err
	}
}

func (hm *HisManager) QueryPy(uid string) (data []map[string]string, err error) {
	// /usr/bin/env python ./main.py --uid 1
	_, output, strerr, _ := runCommand("/usr/bin/env", []string{"python3", "./application/port_his/main.py", "--uid", uid, "--db_path", hm.dbFile})
	if output == "" {
		return nil, errors.New(strerr)
	}

	str := strings.ReplaceAll(output, "'", "\"")
	err = json.Unmarshal([]byte(str), &data)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]string, 0)

	for _, element := range data {
		for k, v := range element {
			if v == "[--请选择--]--请选择--" {
				element[k] = "NA"
			}
		}
		result = append(result, element)
	}

	return result, nil
}

func (hm *HisManager) Query(uid string) (result []map[string]string, err error) {
	// /usr/bin/env python ./main.py --uid 1
	if hm.dbFile == "" {
		return nil, errors.New("db file not set")
	} else if hm.engin == nil {
		return nil, errors.New("db engin not init")
	}

	selector := fmt.Sprintf("%%%s%%", uid)
	data, err := hm.db().Table(hm.table).Where(hm.index, "like", selector).Get()

	if err != nil {
		return nil, err
	}

	for _, element := range data {
		elestr := make(map[string]string, 0)
		for k, v := range element {
			str := v.(string)
			if str == "[--请选择--]--请选择--" {
				elestr[k] = "NA"
			} else {
				elestr[k] = str
			}
		}
		result = append(result, elestr)
	}

	return result, nil
}

func runCommand(commandName string, params []string) (ok bool, output string, outerr string, exitcode int) {
	cmd := exec.Command(commandName, params...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		ok = false
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		ok = false
		return
	}

	cmd.Start()

	stderrReader := bufio.NewReader(stderr)
	stdoutReader := bufio.NewReader(stdout)

	//实时循环读取输出流中的一行内容

	go func() {
		for {
			line, err2 := stderrReader.ReadString('\n')
			if err2 != nil || io.EOF == err2 {
				break
			}
			//fmt.Printf(line)
			outerr += line
		}
	}()

	go func() {
		for {
			line, err2 := stdoutReader.ReadString('\n')
			if err2 != nil || io.EOF == err2 {
				break
			}
			//fmt.Printf(line)
			output += line
		}
	}()

	cmd.Wait()
	//fmt.Println("Run finish")
	exitcode = cmd.ProcessState.ExitCode()
	return true, output, outerr, exitcode
}
