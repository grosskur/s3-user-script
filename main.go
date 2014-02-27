package main

import (
	"bufio"
	"bytes"
	flags "github.com/jessevdk/go-flags"
	awsauth "github.com/smartystreets/go-aws-auth"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"text/template"
)

type Options struct {
	Verbose []bool `short:"v" long:"verbose" description:"verbose output"`
	Url     string `short:"u" long:"url" description:"endpoint URL" default:"https://s3.amazonaws.com"`
	Bucket  string `short:"b" long:"bucket" description:"S3 bucket" required:"true"`
	Key     string `short:"k" long:"key" description:"S3 key" default:"{{.Role}}/user-script"`
}

type Config struct {
	Role string
}

func getIAMRoleList() []string {
	var roles []string
	url := "http://169.254.169.254/latest/meta-data/iam/security-credentials/"

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return roles
	}

	resp, err := client.Do(req)

	if err != nil {
		return roles
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		roles = append(roles, scanner.Text())
	}
	return roles
}

func expandRole(s string, role string) string {
	t := template.New("key template")
	t, err := t.Parse(s)
	if err != nil {
		log.Fatal(err)
	}
	var expandedKey bytes.Buffer
	err = t.Execute(&expandedKey, Config{Role: role})
	if err != nil {
		log.Fatal(err)
	}
	return expandedKey.String()
}

func downloadFile(url string, path string) {
	req, _ := http.NewRequest("GET", url, nil)
	client := &http.Client{}
	awsauth.Sign(req)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode >= 400 {
		log.Fatalf("bad status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	out, err := os.Create(path)
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
}

func main() {
	prog := filepath.Base(os.Args[0])
	log.SetFlags(0)
	log.SetPrefix(prog + ": ")

	var opts Options
	parser := flags.NewParser(&opts, flags.Default)

	_, err := parser.Parse()
	if err != nil {
		os.Exit(2)
	}

	roles := getIAMRoleList()
	if len(roles) > 0 {
		role := roles[0]
		opts.Key = expandRole(opts.Key, role)
	}

	fullUrl := opts.Url + "/" + opts.Bucket + "/" + opts.Key
	if len(opts.Verbose) > 0 {
		log.Printf("downloading: %s\n", fullUrl)
	}
	downloadFile(fullUrl, "user-script")
	err = os.Chmod("user-script", 0700)
	if err != nil {
		log.Fatal(err)
	}

	if len(opts.Verbose) > 0 {
		log.Println("executing: ./user-script")
	}
	err = syscall.Exec("./user-script", []string{"./user-script"}, []string{})
	if err != nil {
		log.Fatal(err)
	}
}
