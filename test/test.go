package main

import (
	"fmt"
	"log"
	"os/exec"
	"reflect"
	"time"

	"github.com/shserver/gopackage/shlog"
)

// import (
// 	"fmt"
// 	"reflect"
// )

// func main() {
// 	type S struct {
// 		F string `species:"gopher" color:"blue"`
// 	}

// 	s := S{}
// 	st := reflect.TypeOf(s)
// 	field := st.Field(0)
// 	fmt.Println(field.Tag.Get("color"), field.Tag.Get("species"))

// }

type Account struct {
	id      int    `shorm:"primary key;auto_increment"`
	role    string `shorm:"varchar(10);not null"`
	user_id string `shorm:"varchar(20);unique;not null"`
	pw      string `shorm:"varchar(64);not null"`
	name    string `shorm:"varchar(10);not null"`
	email   string `shorm:"varchar(40);not null"`
}

func TestFunc() {
}

func main() {
	shlog.InitLogger("")
	// utility.Log("DEBUG", "TEST 111222!!")
	shlog.Logf("INFO", "SignUp Request from client")
}

func GetFieldName(fieldPinter interface{}) {

	t := reflect.TypeOf(fieldPinter)
	fmt.Println(t.Name())
	// fmt.Println(val)

	// fmt.Println(val.Addr())
	// fmt.Println(val.Type())
	// fmt.Println(val.Type().Elem().Name())
}

// type Some struct {
// 	Foo bool
// 	A   string
// 	Bar int
// 	B   string
// }

// func (structPoint *Some) GetFieldName(fieldPinter interface{}) (name string) {

// 	val := reflect.ValueOf(structPoint).Elem()
// 	val2 := reflect.ValueOf(fieldPinter).Elem()

// 	for i := 0; i < val.NumField(); i++ {
// 		valueField := val.Field(i)
// 		if valueField.Addr().Interface() == val2.Addr().Interface() {
// 			return val.Type().Field(i).Name
// 		}
// 	}
// 	return
// }

func terminalExec() {
	start_time := time.Now()
	for i := 0; i < 10; i++ {
		elapsed_time := time.Since(start_time)
		cmd := exec.Command("go", "run", "../client/client.go", elapsed_time.String())
		// cmd.Dir = "../client/"
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(1 * time.Microsecond)
	}
}
