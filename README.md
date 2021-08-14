# Hoagie API Server
This is the repository for the central Hoagie API. It supports authentication using JWT tokens through the Hoagie and CAS system. Currently, it supports the following endpoints:

* `/mail/send` - sends an email using the Hoagie account to the specified listservs and given email content.

More documentation about how to contribute to and work on this repository is in progress.

## Local Development
First, clone the repository with:
```
git clone git@github.com:HoagieClub/api.git
```
You will need to [setup GitHub SSH keys](https://docs.github.com/en/github/authenticating-to-github/connecting-to-github-with-ssh) to successfully run this command. 

Create a new branch that describes your task, for example:
```
git branch -m course-support
```
Now, to run the server locally, you will need a `.env.local` file that contains sensitive information. It will likely be provided to you when the project starts. **Do not share this file with anyone**.

After you have your `.env.local` file, simply run:
```
go get
go run main.go
```
That's it, the server can now be accessed with `http://localhost:8080`.