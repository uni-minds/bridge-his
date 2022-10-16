/*
 * Copyright (c) 2022
 * Author: Xiangyu Liu
 * File: core_test.go
 * Date: 2022/10/15 下午9:09
 */

package manager

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

const DB = "../medical-sys/tmp/database/db_his.sqlite"

func TestHisManager_Query(t *testing.T) {
	var hm HisManager
	var data []map[string]string
	err := hm.Init(DB, "version2", "产妇入院登记号号码")
	if err != nil {
		t.Log(err.Error())
	}
	data, err = hm.Query("2016")
	if err != nil {
		t.Log(err.Error())
	}

	for _, r := range data {

		t.Log(r)
	}
}

func TestHisManager_Init(t *testing.T) {
	csvfile := "./data/main_pacs_server1_studies.csv"
	fp, _ := os.OpenFile(csvfile, os.O_RDONLY, os.ModePerm)
	buf := bufio.NewReader(fp)

	var hm HisManager
	err := hm.Init(DB, "version2", "产妇入院登记号号码")
	if err != nil {
		t.Log(err.Error())
	}

	var c1, c2, cMore, cNone int

	for {
		lined, _, err := buf.ReadLine()
		if err != nil {
			fmt.Println(err.Error())
			break
		}

		str := string(lined)

		d, err := strconv.Atoi(str)
		if err == nil {
			fmt.Println("ignore:", d)
			continue
		}

		strs := strings.Split(str, "-")
		if len(strs) > 1 {
			str = strs[len(strs)-1]
		}
		data, _ := hm.Query(str)

		switch len(data) {
		case 0:
			cNone++
		case 1:
			c1++
		case 2:
			c2++
		default:
			cMore++
		}
	}
	fmt.Printf("1:%d 2:%d 0:%d M: %d", c1, c2, cNone, cMore)
}
