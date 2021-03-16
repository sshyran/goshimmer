#!/bin/bash

now=$(date +"%y%m%d_%H_%M_%S")
mkdir -p csv

echo "Copying csv files from peer_master..."
docker cp peer_master:/go/autopeering-analysis.csv ./csv/$now-peer_master.csv

echo "Copying csv files from peer_master... DONE"
echo "Copied files are located at ./csv"