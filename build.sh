#! /bin/bash

go build -o ./build/hourly .

chmod 777 ./build/hourly
sudo mv ./build/hourly /usr/bin/