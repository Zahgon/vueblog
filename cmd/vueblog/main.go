// Command vueblog mirrors com.markerhub.VueblogApplication.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"github.com/markerhub/vueblog-go/internal/config"
	"github.com/markerhub/vueblog-go/internal/controller"
	"github.com/markerhub/vueblog-go/internal/mapper"
	"github.com/markerhub/vueblog-go/internal/service"
	"github.com/markerhub/vueblog-go/internal/shiro"
	"github.com/markerhub/vueblog-go/internal/util"
)

func main() {
	cfgPath := flag.String("config", "", "path to application.yml (defaults to the committed values)")
	flag.Parse()

	props := config.Defaults()
	if *cfgPath != "" {
		loaded, err := config.Load(*cfgPath)
		if err != nil {
			log.Fatalf("failed to load %s: %v", *cfgPath, err)
		}
		props = loaded
	}

	db, err := sql.Open("mysql", props.Spring.Datasource.URL)
	if err != nil {
		log.Fatalf("failed to open datasource: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Printf("warning: datasource unreachable: %v", err)
	}

	// Spring's component scan, done explicitly.
	jwtUtils := &util.JwtUtils{
		Secret: props.Markerhub.Jwt.Secret,
		Expire: props.Markerhub.Jwt.Expire,
		Header: props.Markerhub.Jwt.Header,
	}
	userService := service.NewUserService(mapper.NewMySQLUserMapper(db))
	blogService := service.NewBlogService(mapper.NewMySQLBlogMapper(db))
	realm := shiro.NewAccountRealm(jwtUtils, userService)

	handler := config.NewRouter(&config.Handlers{
		Account: controller.NewAccountController(userService, jwtUtils),
		Blog:    controller.NewBlogController(blogService),
		User:    controller.NewUserController(userService),
		Filter:  shiro.NewJwtFilter(jwtUtils, realm),
	})

	addr := fmt.Sprintf(":%d", props.Server.Port)
	log.Printf("Started VueblogApplication on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
