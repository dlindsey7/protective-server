# protectiveserver
Web service that listens for repository creations and then protects master

## How to run the service

If you don't already have Go installed, follow the steps for [Getting Started](https://golang.org/doc/install)

Once Go is installed get this project
```
go get github.com/dlindsey7/protectiveserver
```

You'll also need go-github. A library that makes the GitHUb API accessible in Go.
```
go get github.com/google/go-github
```

Navigate to the protectiveserver directory, build and start the service. The port defaults to 3333 but can be anything.
```
protectiveserver go build .
protectiveserver ./protectiveserver -port 3333
```

Next use ngrok to create a thing. Skip installation if you have it already. *Important:* the port must match the one you set earlier.
```
install ngrok
ngrok http 3333
```

## How to test
A successful test will catch the repository creation event from an Organization
1. Create a webhook in the organization using these settings.
   - Payload URL: The ngrok value for http and add '/protect'. For example, ```http://0a8a1f0a.ngrok.io/protect```
   - Secret: "pass"
   - Select "Let me select individual events." -> "Branch and tag creation"
2. Create a new repository in the organization
3. Check the issues in the new repository for the newly created issue
4. Check the master branch protections to see that they've been set
5. The terminal running protectiveserver should show the messages _"Protections updated for: <your repo>/master"_ and _"Successfully created issue"_

## Based on
This project's web server is based on [Learn Go Lesson 5](https://github.com/adnaan/learngo)

Checks for the correct event type based on [Accepting Webhooks with Go](https://groob.io/tutorial/go-github-webhook/)

Other helpful information used to create the correct protections, webhook, and understand the GitHub REST API v3:
- [REST API v3](https://developer.github.com/v3/)
- [Webhooks](https://developer.github.com/webhooks/)
- [Scripting Github](https://git-scm.com/book/en/v2/GitHub-Scripting-GitHub)
