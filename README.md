## w3kipedia

w3kipedia is a try to participate for the [WAM](https://www.wearemillions.online/) hackathon for a better wikipedia on swarm.

It has two components

#### Indexer: 

Indexer can read Wikipedia OpenZip format [snapshots](https://dumps.wikimedia.org/other/kiwix/zim/wikipedia/) create an offline index and
upload content in [swarm](https://www.ethswarm.org/)

#### Server:

Server will start a http-server that can fetch content from swarm and display content in the web browser  

### How to index?
On Ubuntu/Debian :

you need these packages to compile gozim
```
apt-get install git liblzma-dev mercurial build-essential
```

```
cd cmd/indexer
go build
```

On OSX:
```
cd cmd/indexer
CGO_CFLAGS=`pkg-config --cflags liblzma` go build 
```

Docker:
```
docker build -f Dockerfile.indexer --tag w3ki-indexer .
```
#### Help :

```
./indexer --help
Usage of ./indexer:
  -batch string
        bee Postage Stamp ID
  -bee string
        bee API endpoint
  -content
        whether to generate tags  from content for indexing (indexing process will be faster if false)
  -help
        print help
  -index string
        path for the index file
  -offline
        run server offline for listing only
  -proxy
        if Bee endpoint is gateway proxy
  -zim string
        zim file location
```

Docker: 
```
docker run w3ki-indexer -h
```


#### Running :

Binary: 
```
./indexer -index=indexLocation -bee=beeEndpoint -batch=batchID -zim=zimLocation -content=true
```

Docker:
```
docker run 
    -v <PATH_TO_INDEX>:/go/src/github.com/onepeerlabs/w3kipedia/index \
    -v <PATH_TO_ZIM>:/go/src/github.com/onepeerlabs/w3kipedia/<ZIM_FILE_NAME> \
    w3ki-indexer -zim=<ZIM_FILE_NAME> -content=true -index=index -bee=beeEndpoint -batch=batchID
```

### How to serve?

```
cd cmd/server
go build
```

Docker:
```
docker build -f  Dockerfile.server --tag w3ki-server .
```

#### Help :

```
./server -help                                                                                          
Usage of ./server:
  -bee string
        bee API endpoint
  -help
        print help
  -index string
        path for the index file
  -offline
        run server offline for listing only
  -port int
        port to listen to, read HOST env if not specified, default to 8080 otherwise (default -1)
  -proxy
        if Bee endpoint is gateway proxy
```

Docker:
```
docker run w3ki-server -h
```


#### Running :

Binary: 
```
./server -index=indexLocation -bee=beeEndpoint
```

Docker:
```
docker run 
    -p 8080:8080 \
    -v <PATH_TO_INDEX>:/go/src/github.com/onepeerlabs/w3kipedia/index  
    w3ki-server -index=index -bee=beeEndpoint -batch=batchID
```

This will start a local http-serve which will serve wikipedia content on port `:8080`. The index needs to be present locally. 
The content will be served from swarm.

** If index is not present at the given location, server will download the index of top 100 english articles from swarm and serve.

#### How Indexer works:

Indexer uses bleve and boltdb to create index. It is using Article title, url and an array of tags to index each article. 
It also uploads the content of each item in the zim into swarm and saves it in the local index file along with the indexes.
As bleve uses a key-value store, indexer uses relative urls of all the items as key.

- How does it generate tags?

It gets all the words from html content then creates a list of proper nouns, and their occurrences in the article. Takes top 10 words from that list

#### How Server works:

Server lists all the items in the server with "text/html" mimetype. 

For showing the content it checks "GET" request on "/wiki/XXXX". it reads the "XXXX" relative url, then finds the entry for that item in the key-value store,
reads the swarm hash, downloads the content and sends it back as response.

this project [uses code](https://github.com/akhenakh/gozim/blob) that is [MIT licensed](https://github.com/akhenakh/gozim/blob/master/LICENSE)