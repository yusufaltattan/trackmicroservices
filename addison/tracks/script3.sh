#!/bin/sh
ID="Blinding+Lights"
ESCAPED=`perl -e "use URI::Escape; print uri_escape(\"$ID\")"`
RESOURCE=localhost:3000/tracks/$ESCAPED
curl -v -X GET $RESOURCE
