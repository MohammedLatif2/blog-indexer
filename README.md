docker run -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch-oss:6.2.2

docker run -p 5601:5601 docker.elastic.co/kibana/kibana:6.2.2