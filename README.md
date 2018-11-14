# supergloo

## Dev Setup Guide

- After cloning, run `make init` to set up pre-commit githook to enforce Go formatting and imports
- If using IntelliJ/IDEA/GoLand, mark directory `api/external` as Resource Root

## Updating API

- To regenerate API from protos, run `go generate ./...`

## Mesh Config Notes

Go to the [wiki](https://github.com/solo-io/supergloo/wiki) for notes on various service mesh capabilities and configuration. 
