
for i in $(find . | grep .go); do sed -i 's@golang/protobuf@gogo/protobuf@g' "$i"; done
for i in $(find . | grep .go); do sed -i 's@protoc-gen-go@protoc-gen-gogo@g' "$i"; done
for i in $(find . | grep .go); do sed -i 's@pseudomuto@ilackarms@g' "$i"; done
