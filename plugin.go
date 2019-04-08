package main

import (
	"errors"
	"github.com/xanzy/go-gitlab"
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
		Token string
		Asset string
		Name  string
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
	return strings.ReplaceAll(repoLink, fullname, "")
}

//Exec main plugin execution logic ... start here ...
func (p Plugin) Exec() error {

	client := gitlab.NewClient(nil, p.Config.Token)

	if err := client.SetBaseURL(parserBaseUrl(p.Repo.Link, p.Repo.FullName)); err != nil {
		panic(err)
	}

	projectFile, _, err := client.Projects.UploadFile(p.Repo.FullName, p.Config.Asset)

	if err != nil {
		return err
	}

	opts := &gitlab.CreateReleaseOptions{
		Description: &projectFile.Markdown,
	}

	if p.Build.Event == tagEvent {
		opts.TagName = &p.Build.Tag
	} else {
		//todo: accept others events
		return errors.New("event shoud be equals to tag")
	}

	if p.Config.Name != "" {
		opts.Name = &p.Config.Name
	} else {
		opts.Name = String(defaultReleaseName)
	}

	_, _, err = client.Releases.CreateRelease(p.Repo.FullName, opts)

	return err
}
