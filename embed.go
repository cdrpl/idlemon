package main

import _ "embed"

//go:embed robots.txt
var robots string

//go:embed database_up.sql
var upSql string

//go:embed database_down.sql
var downSql string

//go:embed unit_templates.json
var unitTemplatesJson string
