# gtmpl

gtmpl is a little tool to initialize gitlab projects just after creation. It allows to set environment variables for the CI and initialize also files.

```
$ ll $HOME/.gtmpl
total 8
drwxrwxr-x 2 sylvain sylvain 4096 juil. 26 05:27 default
drwxrwxr-x 3 sylvain sylvain 4096 août   2 21:26 docker-scratch
$ cat $HOME/.gtmpl/default/config.yaml

vars:
        - key: foo
          value: bar
          protected: false
          masked: false
          environment_scope: "*"
          variable_type: env_var
        - key: foo2
          value: ba2r
          protected: false
          masked: false
          environment_scope: "*"
          variable_type: env_var
```

```
$ ll  ~/.gtmpl/docker-scratch/data/
total 16
-rw-rw-r-- 1 sylvain sylvain  512 août   3 21:51 Dockerfile
drwxrwxr-x 2 sylvain sylvain 4096 déc.   5  2020 etc
-rw-rw-r-- 1 sylvain sylvain  261 déc.   5  2020 README.md
drwxrwxr-x 2 sylvain sylvain 4096 déc.   5  2020 src
```

```
$ cd .../DEV
$ git clone .../my-new-project
$ cd my-new-project
$ gtmpl -t docker-scratch
```
