# Tiberious
[![Build Status](https://travis-ci.org/JustAnotherOrganization/tiberious.svg?branch=master)](https://travis-ci.org/JustAnotherOrganization/tiberious) [![Issue Count](https://codeclimate.com/github/JustAnotherOrganization/tiberious/badges/issue_count.svg)](https://codeclimate.com/github/JustAnotherOrganization/tiberious)

---
About the name
---
The first existing JIM server (for testing only) was written in Node and was
originally called NodeJIM. Do to the complete boring nature of this name it was
renamed to KirkNode in honor of the Star Trek character James T. Kirk (one of the
more famous Jims that came to mind at the time). Since then KirkNode has been
abandoned for the purpose of writing a more complete JIM server in Go; in keeping
with the naming scheme it was decided that we would call it Tiberious for the T
in James T. Kirk.

---
JIM Protocol
---
The JSON Instant Messaging Protocol is an open protocol alternative to XMPP, IRC
and similar. It is currently still in early stages of development, the general
white paper can be found at https://jim.hackpad.com/.

---
**Requirements**

* Go 1.7 or above
* Godep: `go get github.com/tools/godep`
* Redis: http://redis.io/

**Running**

* Set godep packages: `godep restore`.
* Make sure redis is running.
* Copy example.config.yml to config.yml and make appropriate changes.
* Run or build with go: `go run main.go`.

---

# License   
```
Copyright 2017 Just Another Organization

Licensed under the Apache License, Version 2.0 (the "License");
you may not use these files except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
