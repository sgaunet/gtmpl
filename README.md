# gtmpl

gtmpl is a tool to initialize projects with templates.

* gtmpl will search the templates in $HOME/.gtmpl
* $HOME/.gtmpl should have subfolders representing the name of the templates
* Each template (subfolder) should contain files/folders to copy 

# Usage

```
Usage of gtmpl:
  -d string
        debuglevel : debug/info/warn/error (default "info")
  -f    Overwrite files
  -p string
        Project directory
  -t string
        Template Name
  -v    Get version
```


```
$ ll $HOME/.gtmpl
total 8
drwxrwxr-x 2 sylvain sylvain 4096 juil. 26 05:27 default
drwxrwxr-x 3 sylvain sylvain 4096 août   2 21:26 docker-scratch
```

```
$ cd .../DEV
$ git clone .../my-new-project
$ cd my-new-project
$ gtmpl -t docker-scratch
```

# Major change

Since version 1.0.0, gtmpl is not initializing any gitlab CI variables. It's working only with files.