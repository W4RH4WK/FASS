# Manual (draft)

The Fully Automated Submission System (FASS) is a tool designed to ease the submission process for computer science courses.
It is targeted at a tech-savvy audience, but don't you worry, this manual explains all necessary details.

FASS runs on a Linux server as service, providing a simple HTTP REST interface.
Course / exercise configuration is stored as file and can be modified any time without needing to restart the service.
Users (students) are authenticated via unique tokens, no personal identifiable information should be stored on the server.
A user submits a solution for an exercise by uploading a ZIP archive to the corresponding URL.
The FASS service then invokes the build executable of the corresponding exercise.
Output and exit code are recorded and can be viewed by the user via the HTTP REST interface.

## Installation

We recommend using the latest Ubuntu server OS.

- Install [Go](https://golang.org/) via the system package manager.

      apt install golang

- Add a new user for the FASS service.

      adduser fass
    
- As this new user, install the FASS executable.

      go get github.com/W4RH4WK/FASS/cmd/fass

  The executable is then located in `$HOME/go/bin` which you should add to your path.

- Test the executable.

      fass help

  Since this is the first time you execute FASS, a default configuration file is generated and stored in `$HOME/.config/fass/`.

- Configure the FASS service to your liking by adjusting the values in the config file.

It is recommended to keep the FASS service listening on localhost and use [Nginx](https://www.nginx.com/) as reverse proxy.
This way you can easily add TLS, which is highly recommended.
See Nginx documentation for more information.

- Determine where your course / exercise configuration and submissions will be stored.
  This folder is referred to as *data folder* from here on and is typically located in `$HOME/data` of the FASS user, or `/var/fass`.

- Consider using the provided [service file](fass.service) to start / stop the service via systemd.
  Be sure to adjust the `ExecStart` and `WorkingDirectory` accordingly.
  Use the *data folder* as working directory.

## Course Setup

### Generate Tokens

First, obtain a list of e-mail addresses of the students who registered for your course.
Your facility's management software should provide you this list.
The file (here named `mail.txt`) should contain one e-mail address per line, for instance:

    john.doe@uibk.ac.at
    lo.wang@uibk.ac.at

Using the `fass` executable we can create the *token mapping file* `mapping.json` which associates a randomly generated token with each e-mail address, and hence each student.

    fass token mail.txt

**Important:** The `mail.txt` and `mapping.json` contain PII and should therefore be stored away from the server.
They are not required by the service after the course setup process has been completed.
The instructor uses the `mapping.json` to map a given token to the associated e-mail address.

**Hint:** Add instructor's e-mail addresses to the list so they can test the course / exercise configuration through the same channel that is used by the students.

### Create Course

Each course corresponds to a folder inside the *data folder*.
Furthermore, this *course folder* contains a `course.json` holding some metadata as well as the list of users (tokens) allowed to access the course.
The course number (e.g. `101`) typically serves as an identifier for the course.

Enter your *data folder* and run the following command to create a new course, change the course identifier (here `101`) accordingly.
All tokens from the given mapping file are added to the user list of the course.

    fass course 101 /path/to/mapping.json

This should result in a new folder named `101` containing a `course.json`.
Open this `course.json` file and observe that the tokens from the mapping have been added.
You may want to modify the metadata right away.

### Adding an Exercise

Each exercise corresponds to a folder inside the corresponding *course folder*.
Just pick a meaningful identifier for your exercises (e.g. `ex01`, `ex02`).

Create a new folder `ex01` within your *course folder*.
Upon submission, an executable named `build` within this *exercise folder* will be run.

For testing purposes, copy the following script to *exercise folder* (e.g. `$HOME/data/101/ex01/build`)

```bash
#!/bin/bash

set -eu

echo "====== FASS Build ======"
date
echo "Course:     $FASS_COURSE"
echo "Exercise:   $FASS_EXERCISE"
echo "User:       $FASS_USER"
echo "Submission: $FASS_SUBMISSION"
sha256sum "$FASS_SUBMISSION"
echo
```

Don't forget executable permissions.

    chmod +x build

That's it, submissions (including build output) will be put under `submissions` inside your *exercise folder*.

### Distributing Tokens

Assuming a mail server has been configured in `$HOME/.config/fass/config.json`, the FASS can be used to distribute the generated tokens.
Invoke the following command from your *data directory*.

    fass distribute 101 /path/to/mapping.json

Each student of the course 101 should now receive their token for accessing the service.

**Important:** This is the point where you should remove the `mail.txt` and `mapping.json` file from the server.

## Running

To run service simply issue the following command from your *data directory*.

    fass serve

Alternatively use systemd.

    systemctl start fass.service

## Submission

As the FASS service provides a HTTP REST interface, any tool capable of submitting HTTP requests can be used to interact with the service.
Let's give it a try, substitute `T0K3N`, `localhost:8080`, and `101` according to your configuration.

    curl -H "X-Auth-Token: T0K3N" localhost:8080/api/101

You should now see a list of all exercises of the given course.

Since using `curl` on the commandline can be a bit tedious, we provide [fassup](fassup).
`fassup` is a short script streamlining interaction with the FASS service.
Open `fassup` with a text editor and adjust `HOST`, `COURSE`, and `TOKEN` at the beginning of the file accordingly.

Next we can have a look at the provided commands.

    ./fassup help

Create an ZIP archive and submit it using `fassup`

    touch foo
    zip foo.zip foo
    ./fassup submit ex01 foo.zip

It displays the SHA256 checksum of your file, calculated on the server, and checks your local file for any discrepancies.

After a seconds or so we can query the build output / status.

    ./fassup status ex01

The submission is stored on the server and replaced on consecutive uploads.

    $HOME/data/101/ex01/submissions/T0K3N.zip

## Feedback

Instructors have the option to leave feedback for your submission.

As instructor, put a plaintext file on the server next to the submitted ZIP archive.

    echo "Your solution is awesome!" > "$HOME/data/101/ex01/submissions/T0K3N.feedback"

As student, use `fassup`

    ./fassup feedback ex01

**Note:** It may be helpful to add date and SHA256 checksum of the submission to the feedback so students can be sure that the correct version was reviewed.

## Docker

FASS itself does not specify any rules, restrictions, etc. on how to build / run / test submissions.
It just executes the `build` executable.

In practice you may want to test submissions in an isolated environment.
While not 100% bulletproof, Docker can be used for this purpose.
It is recommended to get yourself familiar with the basics of Docker before continuing.

Typically each course uses a dedicated Docker image containing all that's needed for building / running / testing submissions.

See [example Dockerfile](../example/data/101/Dockerfile).

The `build` executable will now run a Docker instance instead.

See [example `build` executable](../example/data/101/ex01/build).

**Important:** Files located on the host that are made available in the container should be readonly.

We move the actual build / run / test steps to a dedicated script that is executed inside the container.

See [example `build_inside`](../example/data/101/ex01/build_inside).
