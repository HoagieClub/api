# Hoagie API Server
This is the repository for the central Hoagie API. It supports authentication using JWT tokens through the Hoagie and CAS system. Currently, it supports the following endpoints:

* `/mail/send` - sends an email using the Hoagie account to the specified listservs and given email content.

TODO: add more

## Local Development
1. First, clone the repository with the following. You will need to [setup GitHub SSH keys](https://docs.github.com/en/github/authenticating-to-github/connecting-to-github-with-ssh) to successfully run this command. 
```
git clone git@github.com:HoagieClub/api.git
```
2. Run a local MongoDB 6.0 server. [Check out installation instructions here](https://www.mongodb.com/docs/manual/administration/install-community/#std-label-install-community).
3. Rename `.env.local.txt` file to `.env.local`
4. Get the dependencies with:
```
go get
```
5. You can now run the server with
```
go run main.go
```
That's it! The server can now be accessed with `http://localhost:8080`. If there are any issues, you can try running `go run main.go reset` to reset the test database.

## Branches
Create a new branch that describes your task, for example:
```
git branch -m course-support
```