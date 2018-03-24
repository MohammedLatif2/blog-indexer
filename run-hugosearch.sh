#!/bin/sh 


HS_BASEURL=${HS_BASEURL:=http://localhost/}
HS_ELASTIC_BASE=${HS_ELASTIC_BASE:=http://elasticsearch:9200/}
HS_ELASTIC_INDEX=${HS_ELASTIC_INDEX:=hugo}
HS_ELASTIC_TYPE=${HS_ELASTIC_TYPE:=posts}

cat << EOF > config.yml
---
hugo:
  base-url: $HS_BASEURL
  content-root: /data
elastic:
  base: $HS_ELASTIC_BASE
  index: $HS_ELASTIC_INDEX
  type: $HS_ELASTIC_TYPE

EOF

exec ./blog-indexer-linux