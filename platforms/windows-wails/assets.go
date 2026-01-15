package main

import "embed"

//go:embed frontend/index.html frontend/assets/*
var assets embed.FS
