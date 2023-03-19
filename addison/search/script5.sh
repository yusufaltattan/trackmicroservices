#!/bin/sh
ID="_Everybody+(Backstreets+Back)+(Radio+Edit)"
AUDIO=`base64 -i "$ID".wav`
RESOURCE=localhost:3001/search
echo "{ \"Audio\":\"$AUDIO\" }" > input
curl -v -X POST -d @input $RESOURCE
