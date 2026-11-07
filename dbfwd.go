package main

// The k8s backend of dbfwd opens a URL such as k+cnpg://pg.prod/app through
// kubectl port-forward. It is in every build, and it needs kubectl only when
// such a URL is opened.
import _ "github.com/xo/dbfwd/k8s"
