# Context

Context securly stores and conveniently retrieves environment variables in [etcd](https://github.com/coreos/etcd) or [Redis](http://redis.io/).

##Using the CLI.

### Generating a key.

Context uses a single binary file to store both the symetric encryption key and the HMAC secret. This file can be generated using the `key` command. Correct permissions will be set for the resulting file before any data is written to it.

```
$ context key -k /path/to/key
```

The default key location for all commands is `/etc/context/key`.


### Setting and removing values.

Values can be set from the command line using the `set` command. The prompt is password-style, and will not echo your input. Each value is encrypted and stored before input for the next is accepted.

```
$ context set -g myGroup A B C
A=
B=
C=
```

### Retrieving values for execution in context.

Using the `exec` command, you can overwrite values in the current environment with values from the group environment for the execution of a single specified command.

Bellow is non-functional example of attempting to run a debian docker image with Context.

```
$ context exec -g myGroup docker run debian env | sort
HOME=/root
HOSTNAME=b9f394a118f2
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
```

Docker requires that environment variables that are to be passed from the execution environment to the container environment be specified individually. This can be a pain if you have a lot of them.

Context provides a templating mechanism to solve this problem. Using the `-t` flag you can specify that you would like the template token, `{}`, to be replaced in the specified command with a pattern that is resolved for every variable in the group environment.

```
$ context exec -g myGroup -t '-e {}' docker run '{}' debian env | sort
A=1
B=2
C=3
HOME=/root
HOSTNAME=d1b8361cc93c
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
```

Context will not provide the environment variable's value as a substitution for the template token, only its name.



##Design and comparison to other software.

* **No external dependencies.** Key generation and use is handled by Context itself, rather than using PGP or mandating setup using a utility such as openssl, etc.
* **Strong, symetric key encryption by default.** In standard mode, Context uses AES-256 (+ SHA-512 HMAC) to encrypt (and sign) values. 
* **Templated command execution.** To make it easier to wrap underlying commands with information from the environment, Context allows the use of simple templates. 


###Backends and crypters are simple interfaces.

Context makes as few assumptions about the chosen backend as possible, essentially assuming that what's provided is a basic, *binary-safe* key/value store. If a given backend requires content encodings (such as base-64, etc.), the backend-specific package handles any necessary conversions for both reading and writing. Encryption methods are handled much the same way.

There's a straight forward way to add both new backends and crypters that will be user-selectable at run time. We can garauntee that the `std` crypter will remain useable over time by allowing additions of new crypter packages down the line.


## To-dos.

This is by no means complete, but there are a few things that should be added right out of the gate.

* Add a backend for [Consul](https://www.consul.io/).
* Use a more robust parsing mechanism for templates.
* Add the ability to specify a user for the `exec` command.

Definitely contribute if you are so inclined! We'll follow [git flow](http://nvie.com/posts/a-successful-git-branching-model/) with pull requests.