package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	URL          = "ec2-52-37-174-113.us-west-2.compute.amazonaws.com"
	PROTOCOL     = "http://"
	LOCALHOST    = "localhost"
	PORT_NAME    = "8082"
	AUTHPATH     = "/auth/login"
	ACCOUNTPATH  = "/api/accounts"
	PROJECTPATH  = "/api/projects"
	REGISTRYPATH = "/api/registries"
	USERNAME     = "admin"
	PASSWORD     = "shipyard"
)

func setUrl() string {
	url := PROTOCOL + LOCALHOST + ":" + PORT_NAME
	return url
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Authentication struct {
	AuthToken string `json:"auth_token"`
}
type ScriptId struct {
	Id string `json:"id"`
}
type SenderStatistics struct {
	Username     string
	Images       []ScriptImageDetails
	Accounts     []ScriptAccounts
	Projects     []ScriptProjects
	Builds       []ScriptBuilds
	Registries   []ScriptRegistries
	Tests        []ScriptTests
	Results      []ScriptBuildResults
	Repositories []ScriptRepository
}

type ScriptBuilds struct {
	Id        string
	ProjectId string
	TestId    string
	StartTime string
	Status    Status
}
type Status struct {
	Status string
}
type Results struct {
	ResultEntries []string
}
type ScriptProjects struct {
	Name         string
	Author       string
	CreationTime string
	LastRunTime  string
	Status       string
	Images       []*ScriptImageDetails
	Tests        []*ScriptTests
}

type ScriptAccounts struct {
	Id        string
	FirstName string
	LastName  string
	Username  string
	Password  string
	Roles     []string
}

type ScriptImageDetails struct {
	Id             string
	Name           string
	ImageId        string
	Description    string
	Status         string
	RegistryId     string
	Tag            string
	IlmTags        []string
	Location       string
	SkipImageBuild string

	ProjectId string
}

type ScriptRepository struct {
	Name         string
	Tag          string
	FsLayers     []FsLayer
	Signatures   []Signature
	HasProblems  bool
	Message      string
	RegistryUrl  string
	RegistryName string
}

type FsLayer struct {
	BlobSum string
}
type Signature struct {
	Header    Header
	Signature string
	Protected string
}

type Header struct {
	Algorithm string
}
type ScriptRegistries struct {
	Id   string
	Name string
	Addr string
}
type ScriptTests struct {
	Id        string
	ProjectId string
	Provider  Provider
}
type Provider struct {
	providerType string
}
type ScriptBuildResults struct {
	ID            string
	BuildId       string
	ResultEntries []string
}

var projectInfo []ScriptProjects
var registryInfo []ScriptRegistries
var accountInfo []ScriptAccounts
var token string
var credentials Credentials

func setCredentials(u Credentials) Credentials {
	u.Username = USERNAME
	u.Password = PASSWORD
	return u
}

func postAuthentication() []byte {
	path := setUrl() + AUTHPATH
	c := setCredentials(credentials)
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(c)
	result, err := http.Post(path, "application/json; charset=utf-8", b)

	if err != nil {
		panic(err.Error())
	}
	body, err := ioutil.ReadAll(result.Body)
	return body
}

func parseAuthResponse(body []byte) (string, error) {
	var auth Authentication

	err := json.Unmarshal(body, &auth)
	if err != nil {
		fmt.Println(err)
	}
	y := marshalOb(auth)
	split := strings.Split(y, ":")
	authToken := split[1]
	result := authToken[1 : len(authToken)-2]
	return result, err
}

func getIdList(body []byte) []ScriptId {
	var s []ScriptId
	err := json.Unmarshal(body, &s)
	if err != nil {
		fmt.Println(err)
	}
	return s

}

func marshalOb(v interface{}) string {
	vBytes, _ := json.Marshal(v)
	return string(vBytes)
}
func getAuthToken() string {

	body := postAuthentication()
	s, _ := parseAuthResponse(body)
	x := string(s)
	return x
}
func getAccountsfromApi() []byte {
	var body []byte

	token = USERNAME + ":" + getAuthToken()
	url := setUrl() + ACCOUNTPATH
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("X-Access-Token", token)
	response, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
		body, err = ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}
	return body
}

func unmarshalAccounts(body []byte) []ScriptAccounts {
	error := json.Unmarshal(body, &accountInfo)
	if error != nil {
		fmt.Println("error:", error)
	}
	return accountInfo
}

func getProjectsfromApi() []byte {
	var body []byte
	token = USERNAME + ":" + getAuthToken()
	url := setUrl() + PROJECTPATH
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("X-Access-Token", token)
	response, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
		body, err = ioutil.ReadAll(response.Body)

		if err != nil {
			log.Fatal(err)
		}
	}
	return body
}

func getImagesfromApi() []ScriptImageDetails {
	var result []ScriptImageDetails
	var body2 []byte
	token = USERNAME + ":" + getAuthToken()
	url := setUrl() + PROJECTPATH + "/"

	body := getProjectsfromApi()
	s := getIdList(body)
	for _, data := range s {
		projId := url + data.Id + "/images"
		client := &http.Client{}
		req, err := http.NewRequest("GET", projId, nil)
		req.Header.Add("X-Access-Token", token)
		response, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		} else {
			myResult := []ScriptImageDetails{}
			defer response.Body.Close()
			body2, err = ioutil.ReadAll(response.Body)
			json.Unmarshal(body2, &myResult)
			result = append(result, myResult...)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return result
}
func getTestsFromApi() []ScriptTests {
	var body2 []byte
	var result []ScriptTests

	token = USERNAME + ":" + getAuthToken()
	url := setUrl() + PROJECTPATH + "/"
	body := getProjectsfromApi()
	s := getIdList(body)

	for _, data := range s {
		projId := url + data.Id + "/tests"
		client := &http.Client{}
		req, err := http.NewRequest("GET", projId, nil)
		req.Header.Add("X-Access-Token", token)
		response, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		} else {
			myResult := []ScriptTests{}

			defer response.Body.Close()
			body2, err = ioutil.ReadAll(response.Body)
			json.Unmarshal(body2, &myResult)
			result = append(result, myResult...)
			if err != nil {
				log.Fatal(err)
			}

		}
	}
	return result
}

func getBuildsFromApi() []ScriptBuilds {
	var body2 []byte
	var result []ScriptBuilds

	token = USERNAME + ":" + getAuthToken()
	url := setUrl() + PROJECTPATH + "/"
	testsBody := getTestsFromApi()

	for _, data := range testsBody {
		testId := data.Id
		projId := data.ProjectId
		id := url + projId + "/tests/" + testId + "/builds"

		client := &http.Client{}
		req, err := http.NewRequest("GET", id, nil)
		req.Header.Add("X-Access-Token", token)
		response, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		} else {

			myResult := []ScriptBuilds{}

			defer response.Body.Close()
			body2, err = ioutil.ReadAll(response.Body)
			if err != nil {
				log.Fatal(err)
			}
			json.Unmarshal(body2, &myResult)
			result = append(result, myResult...)
		}
	}
	return result
}

func getResultsFromApi() []ScriptBuildResults {
	var body2 []byte
	var result []ScriptBuildResults

	token = USERNAME + ":" + getAuthToken()
	url := setUrl() + PROJECTPATH + "/"
	buildsBody := getBuildsFromApi()
	for _, data := range buildsBody {
		testId := data.TestId
		projId := data.ProjectId
		buildId := data.Id
		id := url + projId + "/tests/" + testId + "/builds/" + buildId + "/results"

		client := &http.Client{}
		req, err := http.NewRequest("GET", id, nil)
		req.Header.Add("X-Access-Token", token)
		response, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		} else {
			myResult := []ScriptBuildResults{}

			defer response.Body.Close()
			body2, err = ioutil.ReadAll(response.Body)
			if err != nil {
				log.Fatal(err)
			}
			json.Unmarshal(body2, &myResult)
			result = append(result, myResult...)
		}
	}
	return result

}

func getRegistriesFromAPi() []ScriptRegistries {
	var body2 []byte

	token = USERNAME + ":" + getAuthToken()
	url := setUrl() + REGISTRYPATH

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("X-Access-Token", token)
	response, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
		body2, err = ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}
	myResult := []ScriptRegistries{}
	json.Unmarshal(body2, &myResult)

	return myResult
}

func getImagesFromRegistriesApi() []ScriptRepository {

	var result []ScriptRepository
	var body2 []byte
	token = USERNAME + ":" + getAuthToken()
	url := setUrl() + "/api/registries/"
	body := getRegistriesFromAPi()

	for _, data := range body {
		projId := url + data.Id + "/repositories"
		client := &http.Client{}
		req, err := http.NewRequest("GET", projId, nil)
		req.Header.Add("X-Access-Token", token)
		response, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		} else {
			myResult := []ScriptRepository{}

			defer response.Body.Close()
			body2, err = ioutil.ReadAll(response.Body)
			json.Unmarshal(body2, &myResult)
			result = append(result, myResult...)

			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return result
}

func unmarshalProjects(body []byte) []ScriptProjects {
	error := json.Unmarshal(body, &projectInfo)
	if error != nil {
		fmt.Println("error:", error)
	}
	return projectInfo
}

func setStatistics(stats SenderStatistics) SenderStatistics {

	uname := "admin"
	img := getImagesfromApi()
	acc := unmarshalAccounts(getAccountsfromApi())
	proj := unmarshalProjects(getProjectsfromApi())
	tst := getTestsFromApi()
	reg := getRegistriesFromAPi()
	bld := getBuildsFromApi()
	res := getResultsFromApi()
	repo := getImagesFromRegistriesApi()
	stats.Username = uname
	stats.Images = img
	stats.Accounts = acc
	stats.Projects = proj
	stats.Tests = tst
	stats.Registries = reg
	stats.Builds = bld
	stats.Results = res
	stats.Repositories = repo
	return stats
}

func postResponse() {
	var stats SenderStatistics
	s := setStatistics(stats)
	b := new(bytes.Buffer)

	json.NewEncoder(b).Encode(s)
	res1, _ := http.Post("https://httpbin.org/post", "application/json; charset=utf-8", b)
	//res1, _ := http.Post("http://16.51.182.155:8080/statistics", "application/json; charset=utf-8", b)
	io.Copy(os.Stdout, res1.Body)
}

func main() {
	postResponse()
}
