# Search Engine (using ElasticSearch)

* [Official Elasticsearch Guide](https://www.elastic.co/guide/en/elasticsearch/reference/7.12/index.html)

Used as repository for persisting searchable data.

```
docker run \
  -d \
  -p 9200:9200 \
  -p 9300:9300 \
  -e "discovery.type=single-node" \
  elasticsearch:7.17.8
```

Add mapping for sorting results **before creating new records**:

```
curl -X PUT -H 'Content-Type: application/json' "http://localhost:9200/tasks" -d '
{
  "mappings": {
    "properties": {
      "id": {
        "type": "keyword"
      },
      "description": {
        "type": "text"
      }
    }
  }
}'
```

This configuration is set automatically when using `docker compose up`, as part of the [`compose.yml`](../compose.yml) file.
