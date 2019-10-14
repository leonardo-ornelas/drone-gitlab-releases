package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"strings"
	"text/template"

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
		Token                  string
		Assets                 []string
		Name                   string
		BaseRepoURL            string
		ReleaseTemplate        string
		AlternativeRepoBaseURL string
		AlternativeRepoName    string
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

func parserBaseURL(repoLink string) string {
	repoLinkURL, _ := url.Parse(repoLink)
	return fmt.Sprintf("%s://%s", repoLinkURL.Scheme, repoLinkURL.Host)
}

func resolveAssetURL(base string, context string) string {
	baseURL, err := url.Parse(base)

	if err != nil {
		log.Fatal(err)
	}

	contextURL, err := url.Parse(context)

	if err != nil {
		log.Fatal(err)
	}

	return baseURL.ResolveReference(contextURL).String()
}

func getReleaseName(p Plugin) *string {
	if !isEmpty(&p.Config.Name) {
		return String(p.Config.Name)
	}
	return String(defaultReleaseName)
}

func normalizePath(file string) ([]string, error) {

	matched, err := filepath.Glob(file)

	if err != nil {
		return nil, err
	}

	if matched == nil {
		return nil, fmt.Errorf("Assets not found: %s", file)
	}

	return matched, nil
}

func plainAssets(rawAssets []string) []string {

	var plainAssetList []string

	for _, rawAsset := range rawAssets {
		_a, err := normalizePath(rawAsset)
		if err == nil {
			plainAssetList = append(plainAssetList, _a...)
		} else {
			log.Println(err)
		}
	}

	return plainAssetList
}

func getReleaseTemplate(p Plugin) string {
	if !isEmpty(&p.Config.ReleaseTemplate) {
		return p.Config.ReleaseTemplate
	}
	return "## Release Notes\n*Commit message*: {{.Commit.Message}}"
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

	baseURL := parserBaseURL(p.Commit.Remote)
	log.Println(fmt.Sprintf("%-20s: %s", "Base URL", baseURL))

	if err := client.SetBaseURL(baseURL); err != nil {
		return err
	}

	repoFullName := p.Repo.FullName
	log.Println(fmt.Sprintf("%-20s: %s", "Repo Full Name", repoFullName))

	repoURL := resolveAssetURL(baseURL, repoFullName)
	log.Println(fmt.Sprintf("%-20s: %s", "Repo URL", repoURL))

	assets := plainAssets(p.Config.Assets)

	if assets == nil {
		return errors.New("Assets list must have values")
	}

	var assetLinks []*gitlab.ReleaseAssetLink

	log.Println("Uploading assets...")

	for _, asset := range assets {

		log.Print(fmt.Sprintf("%-20s: %s", "Uploading asset", asset))

		projectFile, _, err := client.Projects.UploadFile(repoFullName, asset)

		if err != nil {
			log.Println("Error")
			return err
		}

		log.Println(fmt.Sprintf("%-20s: %s", "Done", projectFile.URL))

		assetURL := fmt.Sprintf("%s%s", repoURL, projectFile.URL)
		var ral = gitlab.ReleaseAssetLink{Name: projectFile.Alt, URL: assetURL}
		assetLinks = append(assetLinks, &ral)

	}

	log.Print("Upload successful.")

	if p.Build.Event != tagEvent {
		//todo: accept others events
		return errors.New("Event shoud be TAG")
	}

	rel, _, _ := client.Releases.GetRelease(p.Repo.FullName, p.Build.Tag)

	if rel != nil && !isEmpty(&rel.TagName) {
		_, _, err := client.Releases.DeleteRelease(p.Repo.FullName, p.Build.Tag)
		if err != nil {
			return err
		}
	}

	releaseAssets := gitlab.ReleaseAssets{Links: assetLinks}

	//create release
	opts := &gitlab.CreateReleaseOptions{
		Description: &releaseMessage,
		TagName:     &p.Build.Tag,
		Name:        getReleaseName(p),
		Assets:      &releaseAssets,
	}
	_, _, err = client.Releases.CreateRelease(p.Repo.FullName, opts)
	return err

}
