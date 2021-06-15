package main

import _ "embed"

//go:embed robots.txt
var robots string

//go:embed database_up.sql
var upSql string

//go:embed database_down.sql
var downSql string

//go:embed daily_quests.json
var dailyQuestsJson string

//go:embed resources.json
var resourcesJson string

//go:embed unit_templates.json
var unitTemplatesJson string

//go:embed unit_types.json
var unitTypesJson string
