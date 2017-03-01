package cmd

import (
	"os/exec"

	"github.com/gobuffalo/buffalo/buffalo/cmd/generate"
	"github.com/markbates/gentronics"
)

func newAppGenerator(data gentronics.Data) *gentronics.Generator {
	g := gentronics.New()
	g.Add(gentronics.NewFile("README.md", nREADME))
	g.Add(gentronics.NewFile("main.go", nMain))
	g.Add(gentronics.NewFile(".buffalo.dev.yml", nRefresh))
	g.Add(gentronics.NewFile(".codeclimate.yml", nCodeClimate))

	if data["ciProvider"] == "travis" {
		g.Add(gentronics.NewFile(".travis.yml", nTravis))
	}

	g.Add(gentronics.NewFile("actions/app.go", nApp))
	g.Add(gentronics.NewFile("actions/home.go", nHomeHandler))
	g.Add(gentronics.NewFile("actions/home_test.go", nHomeHandlerTest))
	g.Add(gentronics.NewFile("actions/render.go", nRender))
	g.Add(gentronics.NewFile("grifts/routes.go", nGriftRoutes))
	g.Add(gentronics.NewFile("templates/index.html", nIndexHTML))
	g.Add(gentronics.NewFile("templates/application.html", nApplicationHTML))
	g.Add(gentronics.NewFile(".gitignore", nGitignore))
	g.Add(gentronics.NewCommand(generate.GoGet("github.com/markbates/refresh/...")))
	g.Add(gentronics.NewCommand(generate.GoInstall("github.com/markbates/refresh")))
	g.Add(gentronics.NewCommand(generate.GoGet("github.com/markbates/grift/...")))
	g.Add(gentronics.NewCommand(generate.GoInstall("github.com/markbates/grift")))
	g.Add(gentronics.NewCommand(generate.GoGet("github.com/motemen/gore")))
	g.Add(gentronics.NewCommand(generate.GoInstall("github.com/motemen/gore")))
	g.Add(generate.NewWebpackGenerator(data))
	g.Add(newSodaGenerator())
	g.Add(gentronics.NewCommand(appGoGet()))
	g.Add(generate.Fmt)

	return g
}

func appGoGet() *exec.Cmd {
	appArgs := []string{"get", "-t"}
	if verbose {
		appArgs = append(appArgs, "-v")
	}
	appArgs = append(appArgs, "./...")
	return exec.Command("go", appArgs...)
}

const nREADME = `
# Welcome to Buffalo!

Thank you for chosing Buffalo for your web development needs.

{{#if withPop}}
## Database Setup

It looks like you chose to set up your application using a {{dbType}} database! Fantastic!

The first thing you need to do is open up the "database.yml" file and edit it to use the correct usernames, passwords, hosts, etc... that are appropriate for your environment.

You will also need to make sure that **you** start/install the database of your choice. Buffalo **won't** install and start {{dbType}} for you.

### Create Your Databases

Ok, so you've edited the "database.yml" file and started {{dbType}}, now Buffalo can create the databases in that file for you:

	$ buffalo db create -a
{{/if}}

## Starting the Application

Buffalo ships with a command that will watch your application and automatically rebuild the Go binary and any assets for you. To do that run the "buffalo dev" command:

	$ buffalo dev

If you point your browser to [http://127.0.0.1:3000](http://127.0.0.1:3000) you should see a "Welcome to Buffalo!" page.

**Congratulations!** You now have your Buffalo application up and running.

## What Next?

We recommend you heading over to [http://gobuffalo.io](http://gobuffalo.io) and reviewing all of the great documentation there.

Good luck!

[Powered by Buffalo](http://gobuffalo.io)`

const nMain = `package main

import (
	"fmt"
	"log"
	"net/http"

	"{{actionsPath}}"
	"github.com/gobuffalo/envy"
)

func main() {
	port := envy.Get("PORT", "3000")
	log.Printf("Starting {{name}} on port %s\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), actions.App()))
}

`
const nApp = `package actions

import (
	"github.com/gobuffalo/buffalo"
	{{#if withPop }}
	"github.com/gobuffalo/buffalo/middleware"
	"{{modelsPath}}"
	{{/if}}
	"github.com/gobuffalo/envy"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
func App() *buffalo.App {
	if app == nil {
		app = buffalo.Automatic(buffalo.Options{
			Env: ENV,
			SessionName: "_{{name}}_session",
		})

		{{#if withPop }}
		app.Use(middleware.PopTransaction(models.DB))
		{{/if}}

		app.GET("/", HomeHandler)

		app.ServeFiles("/assets", assetsPath())
	}

	return app
}
`

const nRender = `package actions

import (
	"net/http"

	rice "github.com/GeertJohan/go.rice"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/buffalo/render/resolvers"
	"github.com/gobuffalo/velvet"
)

var r *render.Engine

func init() {
	r = render.New(render.Options{
		HTMLLayout:     "application.html",
		TemplateEngine: velvet.BuffaloRenderer,
		FileResolverFunc: func() resolvers.FileResolver {
			return &resolvers.RiceBox{
				Box: rice.MustFindBox("../templates"),
			}
		},
		Helpers: map[string]interface{}{},
	})
}

func assetsPath() http.FileSystem {
	box := rice.MustFindBox("../public/assets")
	return box.HTTPBox()
}
`

const nHomeHandler = `package actions

import "github.com/gobuffalo/buffalo"

// HomeHandler is a default handler to serve up
// a home page.
func HomeHandler(c buffalo.Context) error {
	return c.Render(200, r.HTML("index.html"))
}
`

const nHomeHandlerTest = `package actions_test

import (
	"testing"

	"{{actionsPath}}"
	"github.com/markbates/willie"
	"github.com/stretchr/testify/require"
)

func Test_HomeHandler(t *testing.T) {
	r := require.New(t)

	w := willie.New(actions.App())
	res := w.Request("/").Get()

	r.Equal(200, res.Code)
	r.Contains(res.Body.String(), "Welcome to Buffalo!")
}
`

const nIndexHTML = `<div class="row">
  <div class="col-md-2">
    <img src="/assets/images/logo.svg" alt="" />
  </div>
  <div class="col-md-10">
    <h1>Welcome to Buffalo! [v{{version}}]</h1>
    <h2>
      <a href="https://github.com/gobuffalo/buffalo"><i class="fa fa-github" aria-hidden="true"></i> https://github.com/gobuffalo/buffalo</a>
    </h2>
    <h2>
      <a href="http://gobuffalo.io"><i class="fa fa-book" aria-hidden="true"></i> Documentation</a>
    </h2>

    <hr>
    <h2>Defined Routes</h2>
    <table class="table table-striped">
      <thead>
        <tr text-align="left">
          <th>METHOD</th>
          <th>PATH</th>
          <th>HANDLER</th>
        </tr>
      </thead>
      <tbody>
        \{{#each routes as |r|}}
        <tr>
          <td>\{{r.Method}}</td>
          <td>\{{r.Path}}</td>
          <td><code>\{{r.HandlerName}}</code></td>
        </tr>
        \{{/each}}
      </tbody>
    </table>
  </div>
</div>

`

const nApplicationHTML = `<html>
<head>
  <meta charset="utf-8">
  <title>Buffalo - {{ titleName }}</title>
  <link rel="stylesheet" href="/assets/application.css" type="text/css" media="all" />
</head>
<body>

  <div class="container">
    \{{ yield }}
  </div>

  <script src="/assets/application.js" type="text/javascript" charset="utf-8"></script>
</body>
</html>
`

const nGitignore = `vendor/
**/*.log
**/*.sqlite
.idea/
bin/
tmp/
node_modules/
.sass-cache/
rice-box.go
public/assets/
{{ name }}
`

const nGriftRoutes = `package grifts

import (
	"os"

	. "github.com/markbates/grift/grift"
	"{{actionsPath}}"
	"github.com/olekukonko/tablewriter"
)

var _ = Add("routes", func(c *Context) error {
	a := actions.App()
	routes := a.Routes()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Method", "Path", "Handler"})
	for _, r := range routes {
		table.Append([]string{r.Method, r.Path, r.HandlerName})
	}
	table.SetCenterSeparator("|")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()
	return nil
})`

const nRefresh = `app_root: .
ignored_folders:
- vendor
- log
- logs
- assets
- public
- grifts
- tmp
- bin
- node_modules
- .sass-cache
included_extensions:
- .go
- .html
- .md
- .js
- .tmpl
build_path: tmp
build_delay: 200ns
binary_name: {{name}}-build
command_flags: []
enable_colors: true
log_name: buffalo
`

const nCodeClimate = `engines:
  fixme:
    enabled: true
  gofmt:
    enabled: true
  golint:
    enabled: true
  govet:
    enabled: true
exclude_paths:
  - grifts/**/*
  - "**/*_test.go"
  - "*_test.go"
  - "**_test.go"
  - logs/*
  - public/*
  - templates/*
ratings:
  paths:
    - "**.go"

`

const nTravis = `language: go
env:
- GO_ENV=test

before_script:
  - psql -c 'create database {{name}}_test;' -U postgres
	- mysql -e 'CREATE DATABASE {{name}}_test;'
  - mkdir -p $TRAVIS_BUILD_DIR/public/assets

go:
  - 1.7.x
  - master

go_import_path: {{ packagePath }}
`
