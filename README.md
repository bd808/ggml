Go Get My Logs
==============

Search and display Logstash formatted logs from an Elasticsearch server.

Usage
-----
```
usage: ggml [<flags>]

Search for logs in a Logstash Elasticsearch index.

Flags:
  --help               Show help.
  -u, --url=http://127.0.0.1:9200
                       Server URL
  -q, --query="*"      Elasticsearch query string
  -f, --filter=FILTER  Search filter
  --start=START        Oldest timestamp to match
  --end=END            Newest timestamp to match
  -d, --duration=15m   Width of timestamp window
  -n, --num=100        Number of results to fetch
  --index-format="logstash-%Y.%m.%d"
                       Index name format
  -o, --output-format="{@timestamp} {host} {type} {level}: {message}"
                       Output format
  --verbose            Enable verbose mode
  --debug              Enable debug mode
  --version            Show application version.
```

Some settings can be provided via environment variables:
* `GGML_URL`: Server URL
* `GGML_OUTPUT`: Output format

License
-------

Go Get My Logs is licensed under the MIT license. See the `LICENSE` file for
more details.
