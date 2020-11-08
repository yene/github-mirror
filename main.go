package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

type apiResponse struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Private      bool   `json:"private"`
	URL          string `json:"url"`
	SSHURL       string `json:"ssh_url"`
	CloneURL     string `json:"clone_url"`
	HasWiki      bool   `json:"has_wiki"`
	HasIssues    bool   `json:"has_issues"`
	HasProjects  bool   `json:"has_projects"`
	HasDownloads bool   `json:"has_downloads"`
	HasPages     bool   `json:"has_pages"`
	HTMLURL      string `json:"html_url"`
}

var backupPath = flag.String("path", "~/github/", "Where the backup will be stored.")
var username = flag.String("user", "", "Github username.")
var secret = flag.String("secret", "", "Github password or access token (https://github.com/settings/tokens)")
var githubPath = flag.String("github-path", "/user", "Path to Github repositories: /users/[username] or /orgs/[orgname]")

const baseURI = "https://api.github.com"

func init() {
	log.SetFlags(log.Lshortfile)
	// log.SetOutput(os.Stdout)
}

func main() {
	flag.Parse()
	if len(*backupPath) == 0 || len(*username) == 0 || len(*secret) == 0 {
		log.Println("Missing Parameters")
		flag.PrintDefaults()
		os.Exit(1)
	}
	path := expandTilde(*backupPath)
	log.Println("Checking folder", path)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			log.Println("Folder does not exist, creating it", path)
			err = os.Mkdir(path, 0777)
			if err != nil {
				log.Fatal("Could not create folder.", err)
			}
		} else {
			log.Fatal("Error when trying to check existing folder", err)
		}
	}

	repos := getAllRepos()
	log.Printf("%d repositories found. \n", len(repos))
	for _, r := range repos {
		backupRepo(r)
	}
	destination := filepath.Join(path, "lastupdated")
	now := time.Now().String()
	err := ioutil.WriteFile(destination, []byte(now+"\n"), 0777)
	if err != nil {
		log.Println("Could not create file", err)
	}
}

func backupRepo(repo apiResponse) {
	log.Println("Handling Repo:", repo.Name)
	getRepo(repo)
	getWiki(repo)
	getIssues(repo)
}

func getRepo(repo apiResponse) {
	destination := filepath.Join(expandTilde(*backupPath), repo.Name, "repo")
	if _, err := os.Stat(destination); os.IsNotExist(err) {
		source := strings.Replace(repo.CloneURL, "https://", "https://"+*username+":"+*secret+"@", 1)
		gitClone(source, destination)
	} else {
		gitFetch(destination)
	}
}

func getWiki(repo apiResponse) {
	if repo.HasWiki {
		destination := filepath.Join(expandTilde(*backupPath), repo.Name, "wiki")
		if _, err := os.Stat(destination); os.IsNotExist(err) {
			wikiURI := strings.Replace(repo.CloneURL, ".git", ".wiki.git", 1)
			source := strings.Replace(wikiURI, "https://", "https://"+*username+":"+*secret+"@", 1)
			gitClone(source, destination)
		} else {
			gitFetch(destination)
		}
	}
}

func getIssues(repo apiResponse) (err error) {
	if repo.HasIssues {
		req, err := http.NewRequest("GET", repo.URL+"/issues", nil)
		if err != nil {
			log.Fatal(err)
		}
		req.SetBasicAuth(*username, *secret)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		destination := filepath.Join(expandTilde(*backupPath), repo.Name, "issues.json")
		err = ioutil.WriteFile(destination, body, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
func getAllRepos() (repos []apiResponse) {
	req, err := http.NewRequest("GET", baseURI+*githubPath+"/repos?per_page=500", nil)
	if err != nil {
		log.Fatal("Couldn't connect to Github", err)
	}
	req.SetBasicAuth(*username, *secret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Request failed", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		content, _ := ioutil.ReadAll(resp.Body)
		log.Fatalln(string(content))
	}

	err = json.NewDecoder(resp.Body).Decode(&repos)
	if err != nil {
		log.Fatal("Received invalid Response, check you're parameters \n", err)
	}
	return repos
}

func expandTilde(p string) string {
	if p[:2] == "~/" {
		usr, err := user.Current()
		if err == nil {
			p = filepath.Join(usr.HomeDir, p[2:])
		}
	}
	return p
}

func gitClone(source string, destination string) {
	args := []string{"clone", source, destination}
	_, err := exec.Command("git", args...).Output()
	if err != nil && !strings.Contains(err.Error(), "128") {
		log.Fatal(err)
	}
}

func gitFetch(folder string) {
	args := []string{"-C", folder, "fetch", "-p"}
	out, err := exec.Command("git", args...).Output()
	if err != nil && !strings.Contains(err.Error(), "128") {
		log.Println("Error when trying to run git fetch -p in:", folder)
		log.Println(string(out))
		return
	}
	args = []string{"-C", folder, "pull"}
	out, err = exec.Command("git", args...).Output()
	if err != nil && !strings.Contains(err.Error(), "128") {
		log.Println("Error when trying to run git fetch -p in:", folder)
		log.Println(string(out))
		return
	}
}
