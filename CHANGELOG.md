## 0.1.3 (unreleased)

* fixed an error that occurred when no keys were set for a group

## 0.1.2

* vendored go-etcd to match the etcd 0.4.6 API

## 0.1.1

* commands fix for incorrect backend constructor
* upgrade to Go 1.4.2

## 0.1.0

* fix for writes against non-primary etcd addresses
	* switched to an etcdctl-style machines list
	* dropped cluster sync
* dropped unix-socket support for the redis backend

## 0.0.2

* fix for empty-string values
* Docker-based builds

## 0.0.1

*Initial release.*
