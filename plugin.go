package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"text/template"

	// "github.com/drone/drone-template-lib/template"
	"github.com/xanzy/go-gitlab"
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
		Token           string
		Assets          []string
		Name            string
		ReleaseTemplate string
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

func isEmpty(s *string) bool {
	return s == nil || len(strings.TrimSpace(*s)) == 0
}

func parserBaseURL(repoLink string, fullname string) string {
	return strings.ReplaceAll(repoLink, fullname+".git", "")
}

func getReleaseName(p Plugin) *string {
	if p.Config.Name != "" {
		return String(p.Config.Name)
	}
	return String(defaultReleaseName)
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

func getReleaseTemplate(p Plugin) string {
	if !isEmpty(&p.Config.ReleaseTemplate) {
		return p.Config.ReleaseTemplate
	}
	return "Commit message: {{.Commit.Message}}"
}

//Exec main plugin execution logic ... start here ...
func (p Plugin) Exec() error {

	var releaseMessageTemplate = getReleaseTemplate(p)

	tmpl, err := template.New("release-message").Parse(releaseMessageTemplate)

	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)

	err = tmpl.Execute(buf, p)

	if err != nil {
		return err
	}

	var releaseMessage = buf.String()

	client := gitlab.NewClient(nil, p.Config.Token)

	if err := client.SetBaseURL(parserBaseURL(p.Commit.Remote, p.Repo.FullName)); err != nil {
		return err
	}

	log.Println(fmt.Sprintf("URL: %s", client.BaseURL().String()))

	var assetLinks []*gitlab.ReleaseAssetLink

	log.Println("Uploading assets...")
	for _, asset := range p.Config.Assets {
		projectFile, _, err := client.Projects.UploadFile(p.Repo.FullName, normalizePath(asset))

		if err != nil {
			return err
		}

		var ral = gitlab.ReleaseAssetLink{Name: projectFile.Alt, URL: projectFile.URL}

		log.Println(fmt.Sprintf("Uploading asset: %s", projectFile.URL))

		assetLinks = append(assetLinks, &ral)

	}

	log.Print("Successful uploaded.")

	if p.Build.Event != tagEvent {
		//todo: accept others events
		return errors.New("Event shoud be TAG")
	}

	rel, _, _ := client.Releases.GetRelease(p.Repo.FullName, p.Build.Tag)

	if !isEmpty(&rel.TagName) {

		// //update release
		// upOpts := gitlab.UpdateReleaseOptions{
		// 	// Description: String(strings.Join(markdowns, "\r\n")),
		// 	Description: String("descricao3"),
		// 	Name:        getReleaseName(p),
		// }

		_, _, err := client.Releases.DeleteRelease(p.Repo.FullName, p.Build.Tag)

		if err != nil {
			return err
		}

		// _, _, err := client.Releases.UpdateRelease(p.Repo.FullName, p.Build.Tag, &upOpts)
		// return err
	}

	releaseAssets := gitlab.ReleaseAssets{Links: assetLinks}

	//create release
	opts := &gitlab.CreateReleaseOptions{
		// Description: String(strings.Join(markdowns, "")),
		Description: &releaseMessage,
		TagName:     &p.Build.Tag,
		Name:        getReleaseName(p),
		Assets:      &releaseAssets,
	}
	_, _, err = client.Releases.CreateRelease(p.Repo.FullName, opts)
	return err

}
