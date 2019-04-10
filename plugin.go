package main

import (
	"errors"
	"github.com/xanzy/go-gitlab"
	"log"
	"path/filepath"
	"strings"
)

const (
	tagEvent           = "tag"
	defaultReleaseName = "Release"
)

type (
	//Repo data
	Repo struct {
		Owner    string
		Name     string
		Link     string
		Avatar   string
		Branch   string
		Private  bool
		Trusted  bool
		FullName string
	}

	//Build data
	Build struct {
		Number   int
		Event    string
		Status   string
		Deploy   string
		Created  int64
		Started  int64
		Finished int64
		Link     string
		Tag      string
	}

	//Commit data
	Commit struct {
		Remote  string
		Sha     string
		Ref     string
		Link    string
		Pull    string
		Branch  string
		Message string
		Author  Author
	}

	//Author data
	Author struct {
		Name   string
		Email  string
		Avatar string
	}

	//Config plugin-specific parameters and secrets
	Config struct {
		Token  string
		Assets []string
		Name   string
	}

	//Plugin main structure
	Plugin struct {
		Repo   Repo
		Build  Build
		Commit Commit
		Config Config
	}
)

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string {
	p := new(string)
	*p = v
	return p
}

func parserBaseUrl(repoLink string, fullname string) string {
	return strings.ReplaceAll(repoLink, fullname+".git", "")
}

func getReleaseName(p Plugin) *string {
	if p.Config.Name != "" {
		return String(p.Config.Name)
	} else {
		return String(defaultReleaseName)
	}
}

func normalizePath(file string) string {

	matched, err := filepath.Glob(file)

	if err != nil {
		panic(err)
	}

	if matched == nil {
		log.Fatal("Assets not found:" + file)
	}

	return matched[0]
}

//Exec main plugin execution logic ... start here ...
func (p Plugin) Exec() error {

	client := gitlab.NewClient(nil, p.Config.Token)

	if err := client.SetBaseURL(parserBaseUrl(p.Commit.Remote, p.Repo.FullName)); err != nil {
		panic(err)
	}

	log.Print("url: " + client.BaseURL().String())

	var markdowns []string
	log.Println("Uploading assets...")
	for _, asset := range p.Config.Assets {
		projectFile, _, err := client.Projects.UploadFile(p.Repo.FullName, normalizePath(asset))

		if err != nil {
			return err
		}

		markdowns = append(markdowns, projectFile.Markdown)
	}

	log.Print("successful")

	if p.Build.Event != tagEvent {
		//todo: accept others events
		return errors.New("event shoud be equals to tag")
	}

	rel, _, _ := client.Releases.GetRelease(p.Repo.FullName, p.Build.Tag)

	if rel != nil && rel.TagName != "" {
		//update release
		upOpts := gitlab.UpdateReleaseOptions{
			Description: String(strings.Join(markdowns, " ")),
			Name:        getReleaseName(p),
		}

		_, _, err := client.Releases.UpdateRelease(p.Repo.FullName, p.Build.Tag, &upOpts)

		return err

	} else {
		//create release
		opts := &gitlab.CreateReleaseOptions{
			Description: String(strings.Join(markdowns, "")),
			TagName:     &p.Build.Tag,
			Name:        getReleaseName(p),
		}

		_, _, err := client.Releases.CreateRelease(p.Repo.FullName, opts)

		return err
	}
}
