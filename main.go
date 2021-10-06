package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"

	"github.com/google/go-github/v39/github"
)

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)
	flag.Parse()
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	name := flag.Arg(0)
	if flag.NArg() == 0 || name == "" {
		log.Println("Organization name must be specified")
		return 1
	}
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		repos, resp, err := github.NewClient(nil).Repositories.ListByOrg(ctx, name, opt)
		if err != nil {
			log.Printf("Failed to list repositories of organization %q", name)
			return 1
		}
		for _, repo := range repos {
			select {
			case <-ctx.Done():
				return 2
			default:
			}
			log.Printf("Cloning %q to %q", *repo.CloneURL, *repo.FullName)
			if err := exec.CommandContext(ctx, "git", "clone", *repo.CloneURL, *repo.FullName).Run(); err != nil {
				log.Printf("Failed to clone %q to %q: %s", *repo.CloneURL, *repo.FullName, err)
				continue
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return 0
}

func main() {
	os.Exit(run())
}
