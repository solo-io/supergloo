#!/bin/bash

# Params:
# namespace name
# label key
# label value

kubectl label ns %s %s=%s --overwrite
