Go Get My Logs
==============

Search and display Logstash formatted logs from an Elasticsearch server.

Usage
-----
```
usage: ggml [<flags>] [<query>]

Search for logs in a Logstash Elasticsearch index.

Flags:
  --help              Show help.
  -u, --url=http://127.0.0.1:9200
                      Server URL
  -m, --must=MUST     Must match
  -x, --must-not=MUST-NOT
                      Must not match
  --start=START       Oldest timestamp to match
  --end=END           Newest timestamp to match
  -d, --duration=15m  Width of timestamp window
  -t, --tail          Tail event stream
  -n, --num=100       Number of results to fetch
  --index-format="logstash-%Y.%m.%d"
                      Index name format
  -o, --output-format="{@timestamp} {host} {type} {level}: {message}"
                      Output format
  --verbose           Enable verbose mode
  --debug             Enable debug mode
  --version           Show application version.

Args:
  [<query>]  Elasticsearch query string
```

Some settings can be provided via environment variables:
* `GGML_URL`: Server URL
* `GGML_OUTPUT`: Output format

Examples:
```
# Custom output format for a specific event type
$ export GGML_OUTPUT="{@timestamp} {level} {channel} {host} {wiki} - {message}"

# Query string
$ ggml type:mediawiki AND NOT channel:api-feature-usage AND host:mw1070

# With must/mustNot filters (filters are cached by Elastcisearch)
$ ggml -m type:mediawiki -x channel:api-feature-usage -m host:mw1070
```

Build
-----
```
export GOPATH=~/golang # Or any other directory you'd like to use
mkdir -p $GOPATH

go get github.com/bd808/ggml

$GOPATH/bin/ggml --help
```

Build Debian package
```
apt-get install dpkg-dev golang-go
git clone https://github.com/bd808/ggml.git
cd ggml
dpkg-buildpackage -b -us -uc
```

License
-------

Go Get My Logs is licensed under the MIT license. See the `LICENSE` file for
more details.
