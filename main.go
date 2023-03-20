package main

import (
	"embed"
	"github.com/gin-gonic/gin"
	"github.com/janatjak/cmsaudit/checker"
	"github.com/janatjak/cmsaudit/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"html/template"
	"net/http"
	"os"
	"sync"
	"time"
)

type ProjectDto struct {
	Name      string
	GitlabUrl string
	WebUrl    string
	Php       string
	Symfony   string
	Cms       string
}

//go:embed templates/*
var f embed.FS

func main() {
	database, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_DSN")), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	database.AutoMigrate(&model.ProjectEntry{})

	token := os.Getenv("TOKEN")

	r := gin.Default()
	templ := template.Must(template.New("").ParseFS(f, "templates/*.html"))
	r.SetHTMLTemplate(templ)

	checkerClient := checker.New(time.Second * 10)

	r.GET("/", func(c *gin.Context) {
		var projects []model.ProjectEntry
		err := database.Find(&projects).Error
		if err != nil {
			panic(err)
		}

		var wg sync.WaitGroup
		wg.Add(len(projects))

		projectDtos := make([]ProjectDto, len(projects))

		for index := range projects {
			go func(index int) {
				project := projects[index]
				result, err := checkerClient.Check(project.WebUrl + "/api/_cms-audit?token=" + token)
				if err != nil {
					projectDtos[index] = ProjectDto{
						Name:      project.Name,
						GitlabUrl: project.GitlabUrl,
						WebUrl:    project.WebUrl,
						Php:       "?",
						Symfony:   "?",
						Cms:       "?",
					}
				} else {
					projectDtos[index] = ProjectDto{
						Name:      project.Name,
						GitlabUrl: project.GitlabUrl,
						WebUrl:    project.WebUrl,
						Php:       result.Php,
						Symfony:   result.Packages[0].Versions["symfony/framework-bundle"].Version,
						Cms:       result.Packages[0].Versions["uxf/cms"].Version,
					}
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
