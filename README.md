This is a fork of CoreOS's [cloud-init repo](https://github.com/coreos/coreos-cloudinit).  Documentation might not be exactly up to date with the current code's functionality.

## Configuration with cloud-config

A subset of the [official cloud-config spec][official-cloud-config] is implemented by cloud-init.
Additionally, kip/itzo specific options are implemented to support a lightweight boot configuration for cloud instances


All supported cloud-config parameters are [documented here][all-cloud-config]. 

[official-cloud-config]: http://cloudinit.readthedocs.org/en/latest/topics/format.html#cloud-config-data
[all-cloud-config]: https://github.com/elotl/cloud-init/tree/master/Documentation/cloud-config.md

The following is an example cloud-config document:

```
#cloud-config

users:
  - name: core
    passwd: $1$allJZawX$00S5T756I5PGdQga5qhqv1

write_files:
  - path: /etc/resolv.conf
    content: |
        nameserver 192.0.2.2
        nameserver 192.0.2.3
```

## Executing a Script

cloud-init supports a custom section for a user supplied script:

```
runscript: |
  #!/bin/bash
  echo 'Hello, world!'
```

## user-data Field Substitution

cloud-init will replace the following set of tokens in your user-data with system-generated values.

| Token         | Description |
| ------------- | ----------- |
| $public_ipv4  | Public IPv4 address of machine |
| $private_ipv4 | Private IPv4 address of machine |

These values are determined based on the given provider on which your machine is running.
