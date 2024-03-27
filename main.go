package main

import (
	"embed"
	"github.com/gin-gonic/gin"
	"github.com/janatjak/cmsaudit/apichecker"
	"github.com/janatjak/cmsaudit/model"
	"github.com/janatjak/cmsaudit/nodechecker"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

type ProjectDto struct {
	Name      string
	GitlabUrl string
	WebUrl    string
	Server    string
	// API
	ApiPhp     string
	ApiSymfony string
	ApiCms     string
	// Node
	WebNode   string
	WebNextJS string
	WebUI     string
	//
	AdminNode   string
	AdminNextJS string
	AdminUI     string
}

//go:embed templates/*
var f embed.FS

func main() {
	user := os.Getenv("AUTH_USER")
	password := os.Getenv("AUTH_PASSWORD")
	database, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_DSN")), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	database.AutoMigrate(&model.ProjectEntry{})

	token := os.Getenv("TOKEN")

	r := gin.Default()
	templ := template.Must(template.New("").ParseFS(f, "templates/*.html"))
	r.SetHTMLTemplate(templ)

	apiCheckerClient := apichecker.New(time.Second * 10)
	nodeCheckerClient := nodechecker.New(time.Second * 10)

	r.GET("/", func(c *gin.Context) {
		var projects []model.ProjectEntry
		err := database.Order("name").Find(&projects).Error
		if err != nil {
			panic(err)
		}

		var wg sync.WaitGroup
		wg.Add(len(projects))

		projectDtos := make([]ProjectDto, len(projects))

		for index := range projects {
			go func(index int) {
				project := projects[index]
				u, err := url.Parse(project.WebUrl)
				if err != nil {
					panic(err)
				}

				baseUrl := u.Scheme + "://" + user + ":" + password + "@" + u.Host
				resultApi, _ := apiCheckerClient.Check(baseUrl + "/api/_cms-audit?token=" + token)
				resultNodeWeb, _ := nodeCheckerClient.Check(baseUrl + "/uxf-audit.json")
				resultNodeAdmin, _ := nodeCheckerClient.Check(baseUrl + "/admin/uxf-audit.json")

				server := "?"
				apiPhp := "?"
				apiSymfony := "?"
				apiCms := "?"
				if resultApi != nil {
					server = resultApi.Server
					apiPhp = resultApi.Php
					apiSymfony = resultApi.Packages[0].Versions["symfony/framework-bundle"].Version
					apiCms = resultApi.Packages[0].Versions["uxf/cms"].Version
				}

				webNode := "?"
				webNextJS := "?"
				webUI := "?"
				if resultNodeWeb != nil {
					webNode = resultNodeWeb.Node
					webNextJS = resultNodeWeb.Next
					webUI = resultNodeWeb.Packages["@uxf/ui"].Version
				}

				adminNode := "?"
				adminNextJS := "?"
				adminUI := "?"
				if resultNodeAdmin != nil {
					adminNode = resultNodeAdmin.Node
					adminNextJS = resultNodeAdmin.Next
					adminUI = resultNodeAdmin.Packages["@uxf/ui"].Version
				}

				projectDtos[index] = ProjectDto{
					Name:        project.Name,
					GitlabUrl:   project.GitlabUrl,
					WebUrl:      project.WebUrl,
					Server:      server,
					ApiPhp:      apiPhp,
					ApiSymfony:  apiSymfony,
					ApiCms:      apiCms,
					WebNode:     webNode,
					WebNextJS:   webNextJS,
					WebUI:       webUI,
					AdminNode:   adminNode,
					AdminNextJS: adminNextJS,
					AdminUI:     adminUI,
				}

				wg.Done()
			}(index)
		}
		wg.Wait()

		c.HTML(http.StatusOK, "index.html", gin.H{
			"projects": projectDtos,
		})
	})

	r.Run()
}
