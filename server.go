package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"github.com/google/go-github/github"
	"context"

	"strings"
	"log"
)

var secretKey = "pass"

type server struct {
	httpServer *http.Server
	listener   net.Listener
}

func (s *server) listenAndServe() error {

	listener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return err
	}
	s.listener = listener
	go s.httpServer.Serve(s.listener)
	fmt.Println("Server now listening")
	return nil

}

func (s *server) shutdown() error {

	if s.listener != nil {
		err := s.listener.Close()
		s.listener = nil
		if err != nil {
			return err
		}
	}
	fmt.Println("Shutting down server")
	return nil

}

func newServer(port string) *server {

	//Custom HTTP Handler
	handle := func(w http.ResponseWriter, r *http.Request) {
		// TODO: Add username password input
		tp := github.BasicAuthTransport{
			Username: strings.TrimSpace(""),
			Password: strings.TrimSpace(""),
		}

		client := github.NewClient(tp.Client())

		// Every webhook needs to have their secret passcode set to this secretKey value
		payload, err := github.ValidatePayload(r, []byte(secretKey))
		if err != nil {
			log.Printf("error validating request body: err=%s\n", err)
			return
		}
		defer r.Body.Close()

		event, err := github.ParseWebHook(github.WebHookType(r), payload)
		if err != nil {
			log.Printf("could not parse webhook: err=%s\n", err)
			return
		}

		//Check event type
		switch e := event.(type) {
		case *github.RepositoryEvent:
			fmt.Print("Received a repository event\n")
			// Leaving this code here because it might be useful with some changes.
			// The issue is that the repo creation isn't the master branch creation. If this event is resent then the
			// master branch exists and this will update the protections. That's not really the use case though.

			/*if e.Action != nil && *e.Action == "created" {
				protectionRequest := &github.ProtectionRequest{
					RequiredStatusChecks: &github.RequiredStatusChecks{
						Strict:   true,
						Contexts: []string{"continuous-integration"},
					},
				}

				fmt.Print(*e.Repo.Owner.Login)

				_, _, err = client.Repositories.GetBranch(context.Background(), *e.Repo.Owner.Login, *e.Repo.Name, "master")

				if err != nil {
					log.Printf("Unable to find the master branch. Please make sure it was created.")
				} else {
					_, _, err = client.Repositories.UpdateBranchProtection(context.Background(), *e.Repo.Owner.Login, *e.Repo.Name, "master", protectionRequest)
					if err != nil {
						log.Printf("Repositories.UpdateBranchProtection() returned error: %v", err)
					} else {
						log.Printf("Protections updated for: %s/master", *e.Repo.Name)
					}

					issue := &github.IssueRequest{
						Title:    github.String("New protections added to master"),
						Body:     github.String("New protections were added to master. Please verify that they're correct @dlindsey7. The new protection is a status check that requires branches to be up to date before merging."),
					}

					_, _, err = client.Issues.Create(context.Background(), *e.Repo.Owner.Login, *e.Repo.Name, issue)
					if err != nil {
						log.Printf("Issues.Create() returned error: %v", err)
					} else {
						log.Print("Successfully created issue")
					}
				}
			}*/

		case *github.CreateEvent:
			// We're specifically looking for the creation of the master branch
			if e.Ref != nil && *e.Ref == "master" {
				// Setup some basic protections
				protectionRequest := &github.ProtectionRequest{
					RequiredStatusChecks: &github.RequiredStatusChecks{
						Strict:   true,
						Contexts: []string{"continuous-integration"},
					},
				}

				// Check that the master branch exists
				_, _, err = client.Repositories.GetBranch(context.Background(), *e.Repo.Owner.Login, *e.Repo.Name, "master")

				if err != nil {
					log.Printf("Unable to find the master branch. Please make sure it was created.")
				} else {
					// Update/Add protections to master
					_, _, err = client.Repositories.UpdateBranchProtection(context.Background(), *e.Repo.Owner.Login, *e.Repo.Name, "master", protectionRequest)
					if err != nil {
						log.Printf("Repositories.UpdateBranchProtection() returned error: %v", err)
					} else {
						log.Printf("Protections updated for: %s/master", *e.Repo.Name)

						// If the protections were added create an issue to notify @dlindsey7
						issue := &github.IssueRequest{
							Title: github.String("New protections added to master"),
							Body:  github.String("New protections were added to master. Please verify that they're correct @dlindsey7. The new protection is a status check that requires branches to be up to date before merging."),
						}

						_, _, err = client.Issues.Create(context.Background(), *e.Repo.Owner.Login, *e.Repo.Name, issue)
						if err != nil {
							log.Printf("Issues.Create() returned error: %v", err)
						} else {
							log.Print("Successfully created issue")
						}
					}
				}
			}

		default:
			log.Printf("unknown event type %s\n", github.WebHookType(r))
			return
		}

	}
	mux := http.NewServeMux()
	mux.HandleFunc("/protect", handle)

	httpServer := &http.Server{Addr: ":" + port, Handler: mux}
	return &server{httpServer: httpServer}

}

func main() {
	var port string
	flag.StringVar(&port, "port", "3333", "./protectiveserver -port 3333")
	flag.Parse()

	// Channels to listen for signal and set down to trigger the goroutine exit.
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	moveAlong := func() {
		fmt.Println("Shutting down...")
	}

	// Sets up server to listen and defers the "shutting down" message until server.shutdown().
	server := newServer(port)
	server.listenAndServe()
	defer moveAlong()

	// go routine that blocks itself until it receives a signal, then it exits.
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		server.shutdown()
		done <- true
	}()

	fmt.Println("Ctrl-C to interrupt...")
	<-done
	fmt.Println("Exiting...")
}
