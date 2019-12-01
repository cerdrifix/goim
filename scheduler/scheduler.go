package main

import (
	"fmt"
	"os"
)

var (
	DbHostname = os.Getenv("PGHOST")
	DbPort     = os.Getenv("PGPORT")
	DbName     = os.Getenv("PGDATABASE")
	DbUsername = os.Getenv("PGUSER")
	DbPassword = os.Getenv("PGPASSWORD")
)

func main() {
	fmt.Println("Scheduler started")
}
