# gtmpl

gtmpl is a little tool to initialize gitlab projects just after creation. It allows to set environment variables for the CI and initialises also files.

ll $HOME/.gtmpl
total 8
drwxrwxr-x 2 sylvain sylvain 4096 juil. 26 05:27 default
drwxrwxr-x 3 sylvain sylvain 4096 août   2 21:26 docker
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

ll  $HOME/.gtmpl/docker/data/

cd .../a-gitlab-project
gtmpl -t docker
