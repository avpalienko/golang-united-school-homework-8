package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Arguments map[string]string
type Item struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}
type fileDb []Item

const filePerm = 0644

var (
	errOper            = errors.New("-operation flag has to be specified")
	errFileName        = errors.New("-fileName flag has to be specified")
	msgErrOperNotFound = "Operation %v not allowed!"
	errItem            = errors.New("-item flag has to be specified")
	msgErrItemExists   = "Item with id %v already exists"
	errId              = errors.New("-id flag has to be specified")
	msgErrIdNotFound   = "Item with id %v not found"
)

func Perform(args Arguments, writer io.Writer) error {
	oper, ok := args["operation"]
	if oper == "" || !ok {
		return errOper
	}
	if args["fileName"] == "" {
		return errFileName
	}
	if !isValidOperation(oper) {
		return fmt.Errorf(msgErrOperNotFound, oper)
	}

	switch oper {
	case "list":
		input, err := ioutil.ReadFile(args["fileName"])
		if err != nil {
			fmt.Println(err)
			return err
		}
		writer.Write(input)
	case "add":
		return addItem(args["fileName"], args["item"], writer)
	case "remove":
		return removeId(args["fileName"], args["id"], writer)
	case "findById":
		return findId(args["fileName"], args["id"], writer)
	}
	return nil
}

func addItem(fn string, item string, writer io.Writer) error {
	if item == "" {
		return errItem
	}
	var itm Item
	err := json.Unmarshal([]byte(item), &itm)
	if err != nil {
		return err
	}
	var db fileDb
	err = db.loadDb(fn)
	if err != nil {
		return err
	}

	str, err := db.add(itm)
	if err != nil {
		return err
	}
	writer.Write([]byte(str))

	if err = db.saveDb(fn); err != nil {
		return err
	}

	return nil
}

func removeId(fn string, id string, writer io.Writer) error {
	if id == "" {
		return errId
	}
	var db fileDb
	err := db.loadDb(fn)
	if err != nil {
		return err
	}

	str, err := db.del(id)
	if err != nil {
		return err
	}
	writer.Write([]byte(str))

	if err = db.saveDb(fn); err != nil {
		return err
	}

	return nil
}

func findId(fn string, id string, writer io.Writer) error {
	if id == "" {
		return errId
	}
	var db fileDb
	err := db.loadDb(fn)
	if err != nil {
		return err
	}

	var str string
	item, idx := db.findById(id)
	if idx > 0 {
		var content []byte
		if content, err = json.Marshal(item); err != nil {
			return err
		}
		str = string(content)
	}

	writer.Write([]byte(str))

	return nil
}

func (db *fileDb) add(itm Item) (string, error) {
	if _, idx := db.findById(itm.Id); idx >= 0 {
		return fmt.Sprintf(msgErrItemExists, itm.Id), nil
	}
	*db = append(*db, itm)
	return "", nil
}

func (db *fileDb) del(id string) (string, error) {
	_, idx := db.findById(id)
	if idx < 0 {
		return fmt.Sprintf(msgErrIdNotFound, id), nil
	}
	tmp := *db
	*db = append(tmp[:idx], tmp[idx+1:]...)
	return "", nil
}

func (db fileDb) findById(id string) (Item, int) {
	for i := range db {
		if db[i].Id == id {
			return db[i], i
		}
	}
	return Item{}, -1
}

func (db *fileDb) loadDb(fn string) error {
	file, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE, filePerm)
	defer file.Close()
	if err != nil {
		return err
	}

	content, err := ioutil.ReadFile(fn)

	if err != nil {
		return err
	}
	if len(content) > 0 {
		if err = json.Unmarshal(content, db); err != nil {
			return err
		}
	}
	return nil
}

func (db *fileDb) saveDb(fn string) (err error) {
	var content []byte
	if content, err = json.Marshal(db); err != nil {
		return err
	}
	err = ioutil.WriteFile(fn, content, filePerm)
	if err != nil {
		return err
	}
	return nil
}

func parseArgs() Arguments {
	var (
		op  = flag.String("operation", "", "operation")
		itm = flag.String("item", "", "item value in json")
		fn  = flag.String("fileName", "", "name of the file")
		id  = flag.String("id", "", "id")
	)
	args := Arguments{
		"operation": "",
		"item":      "",
		"fileName":  "",
		"id":        "",
	}
	flag.Parse()
	args["operation"] = *op
	args["item"] = *itm
	args["fileName"] = *fn
	args["id"] = *id

	return args
}

func isValidOperation(operation string) bool {
	switch operation {
	case
		"add",
		"list",
		"findById",
		"remove":
		return true
	}
	return false
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
