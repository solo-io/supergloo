find . -name '*_test.go' | while read -r f; do \
  numSuites=$(find "$(dirname "$f")" -name '*_suite_test.go' -maxdepth 1 | wc -l | tr -d '[:space:]')

  if [[ "$numSuites" != "1" ]]; then
    echo "Directory $(dirname "$f") is missing a _suite_test.go file"
    exit 1
  fi
done
