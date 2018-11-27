package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	zBarPath *string
)

func init() {
	zBarPath = flag.String("zbar", "C:\\Program Files (x86)\\ZBar\\bin\\zbarcam.exe", "path to zbarcam.exe")
}

func main() {
	cmd := exec.Command(*zBarPath)

	var out bytes.Buffer
	cmd.Stdout = &out

	db, err := sql.Open("sqlite3", "./students.db")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	go cmd.Start()

	for {
		data, err := out.ReadString('\n')

		if err == nil {
			parts := strings.Split(data[:len(data)-2], ":")
			if len(parts) == 2 {

				if parts[0] == "QR-Code" {
					rowName := db.QueryRow("select name from badge where badge_id=?", parts[1])

					var name string
					err := rowName.Scan(&name)

					if err == nil {

						rowSignlog := db.QueryRow("select event from sign_log ORDER BY id DESC LIMIT 1")

						var event string
						err := rowSignlog.Scan(&event)

						if err == nil && event == "out" {
							fmt.Printf("%s Signed In at %s\n", name, time.Now().Format(time.Kitchen))
							db.Exec("insert into sign_log(badge_id, event_timestamp, event) values (?, ?, 'in')", parts[1], strconv.FormatInt(time.Now().Unix(), 10))
						} else {
							fmt.Printf("%s Signed Out at %s\n", name, time.Now().Format(time.Kitchen))
							db.Exec("insert into sign_log(badge_id, event_timestamp, event) values (?, ?, 'out')", parts[1], strconv.FormatInt(time.Now().Unix(), 10))
						}

					} else {
						fmt.Println("Error reading badge, please try again.")
					}

				}
			}
		} else {
			fmt.Println("Error reading badge, please try again.")
		}

	}
}
