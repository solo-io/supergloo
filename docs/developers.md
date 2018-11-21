## Development Guide

- After cloning, run `make init` to set up pre-commit githook to enforce Go formatting and imports
- If using IntelliJ/IDEA/GoLand, mark directory `api/external` as Resource Root

### Updating API

The API is auto-generated from the protos in the api directory. After making changes, 
run `make generated-code -B` to re-generate. 

