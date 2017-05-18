# cloud-initserver
Serve cloud-init data over http

## Usage cloud-init-server

```shell
Usage of ./cloud-init-server:
  -bind string
        Address to bind on defaults to :80 (default ":80")
  -config string
        Path to put cloud-init config files in (default "/etc/cloud-init")
```

## Configuration structure

The cloud-init data is stored in json files containing the two subobjects, these files are stored under a file called after the macaddress of the caller.
Example structure

```
/etc/cloud-init/
/etc/cloud-init/a6:df:6b:76:78:f7
/etc/cloud-init/6a:90:49:79:62:50
```

Where `/etc/cloud-init/6a:90:49:79:62:50` contains:

```
{
    "meta-data":
        {
            "local-hostname": "myhostname"
        },
    "user-data":
        {
            "users": [
                {
                    "name": "myusername",
                    "plain_test_passwd": "mypassword",
                    "shell": "/bin/bash",
                    "sudo": "ALL=(ALL) ALL"
                }
            ]
        }
}
