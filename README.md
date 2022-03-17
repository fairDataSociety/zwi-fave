## w3kipedia

w3kipedia has two components 

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

#### Running :
```
./indexer -index=indexLocation -bee=beeEndpoint -batch=batchID -zim=zimLocation -content=true
```

### How to serve?

```
cd cmd/server
go build
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

#### Running :
```
./server -index=indexLocation -bee=beeEndpoint
```

This will start a local http-serve which will serve wikipedia content on port `:8080`. The index needs to be present locally. 
The content will be served from swarm.

** If index is not present at the given location, server will download the index of top 100 english articles from swarm and serve.

this project [uses code](https://github.com/akhenakh/gozim/blob) that is [MIT licensed](https://github.com/akhenakh/gozim/blob/master/LICENSE)