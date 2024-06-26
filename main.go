package main

import (
	"fmt"
	"log"
	"syscall"

	"github.com/RealistikOsu/RealistikAPI/app"
	"github.com/RealistikOsu/RealistikAPI/beatmapget"
	"github.com/RealistikOsu/RealistikAPI/common"
	"github.com/valyala/fasthttp"

	// Golint pls dont break balls
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/serenize/snaker"
	"gopkg.in/thehowl/go-osuapi.v1"
)

// Version is the git hash of the application. Do not edit. This is
// automatically set using -ldflags during build time.
var Version string

func init() {
	log.SetFlags(log.Ltime)
	log.SetPrefix(fmt.Sprintf("%d|", syscall.Getpid()))
}

func main() {
	fmt.Print("The RealistikOsu API! Because how much more can I break?")
	if Version != "" {
		fmt.Print("; git commit hash: ", Version)
	}
	fmt.Println()

	settings := common.LoadSettings()

	dns := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4,utf8&collation=utf8mb4_general_ci",
		settings.DB_USER,
		settings.DB_PASS,
		settings.DB_HOST,
		settings.DB_PORT,
		settings.DB_NAME,
	)

	db, err := sqlx.Open(settings.DB_SCHEME, dns)
	if err != nil {
		log.Fatalln(err)
	}

	db.MapperFunc(func(s string) string {
		if x, ok := commonClusterfucks[s]; ok {
			return x
		}
		return snaker.CamelToSnake(s)
	})

	beatmapget.Client = osuapi.NewClient(settings.OSU_API_KEY)
	beatmapget.DB = db

	engine := app.Start(db)

	err = fasthttp.ListenAndServe(fmt.Sprintf(":%d", settings.APP_PORT), engine.Handler)
	if err != nil {
		log.Fatalln(err)
	}
}

var commonClusterfucks = map[string]string{
	"RegisteredOn": "register_datetime",
	"UsernameAKA":  "username_aka",
	"BeatmapMD5":   "beatmap_md5",
	"Count300":     "300_count",
	"Count100":     "100_count",
	"Count50":      "50_count",
	"CountGeki":    "gekis_count",
	"CountKatu":    "katus_count",
	"CountMiss":    "misses_count",
	"PP":           "pp",
}
