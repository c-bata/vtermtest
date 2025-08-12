#!/bin/bash

echo "=== vtermtest-cli Examples ==="

# Build
go build -o vtermtest-cli || exit 1

# Basic
echo "1. Basic:"
./vtermtest-cli --command "echo Hello"

echo -e "\n2. Interactive:"
./vtermtest-cli --command "sh -c 'read x; echo Got: \$x'" --keys "test<Enter>"

echo -e "\n3. With WaitFor and custom delimiters:"
./vtermtest-cli --command "sh -c 'sleep 0.5; echo Ready'" --keys "[WaitFor Ready]" --delimiter "[]"

echo -e "\n4. File output:"
./vtermtest-cli --command "date" --output /tmp/screen.txt
echo "File: $(cat /tmp/screen.txt)"

echo -e "\nDone!"