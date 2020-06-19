# ReMAP Backend

A simple backend solution for the Remote Monitoring Application in Psychiatry (ReMAP) App written in Go.

## Specification

* User authentication, anonymous if possible, username/password otherwise
  * username must be unique, backend sends an error if it already exists
* POST arbitrary JSON events
* Upload audio files (10MB max)
* GET list of active surveys
