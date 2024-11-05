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

type Template struct {
	Projects    []ProjectDto
	WebPackages []string
}

type ProjectDto struct {
	Name      string
	GitlabUrl string
	WebUrl    string
	Branch    string
	Api       *apichecker.Audit
	Web       *nodechecker.Audit
	Admin     *nodechecker.Audit
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
	templ := template.Must(template.New("").ParseFS(f, "templates/*.gohtml"))
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

				projectDtos[index] = ProjectDto{
					Name:      project.Name,
					GitlabUrl: project.GitlabUrl,
					WebUrl:    project.WebUrl,
					Branch:    project.Branch,
					Api:       resultApi,
					Web:       resultNodeWeb,
					Admin:     resultNodeAdmin,
				}

				wg.Done()
			}(index)
		}
		wg.Wait()

		c.HTML(http.StatusOK, "index.gohtml", Template{
			Projects: projectDtos,
			WebPackages: []string{
				"@uxf/analytics",
				"@uxf/core",
				"@uxf/data-grid",
				"@uxf/datepicker",
				"@uxf/form",
				"@uxf/localize",
				"@uxf/router",
				"@uxf/styles",
				"@uxf/translations",
				"@uxf/wysiwyg",
				"@uxf/eslint-config",
				"@uxf/icons-generator",
				"@uxf/resizer",
				"@uxf/scripts",
			},
		})
	})

	r.Run()
}
