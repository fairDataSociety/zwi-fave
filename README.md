## w3kipedia

[w3kipedia](https://github.com/onepeerlabs/w3kipedia) was originally a try to participate for the [WAM](https://www.wearemillions.online/) hackathon for a better wikipedia on swarm.

Now [w3kipedia-fave](https://github.com/onepeerlabs/w3kipedia-fave) is a modified version of w3kipedia that uses FaVe to store, index and search the content.

It has two components

#### Uploader: 

[Uploader](./cmd/uploader/README.md) can read Wikipedia OpenZip format [snapshots](https://dumps.wikimedia.org/other/kiwix/zim/wikipedia/) and upload data to FaVe, ultimately storing content on [swarm](https://www.ethswarm.org/).

#### Server:

Server will start a http-server that can fetch content from FaVe and display them in the web browser.  



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
  -fave string
        FaVe API endpoint ("http://localhost:1234/v1")
  -collection string
        Collection name to store on FaVe
  -port int
        port to listen to, read HOST env if not specified, default to 8080 otherwise (default -1)
```

Docker:
```
docker run w3ki-server -h
```


#### Running :

Binary: 
```
./server -fave=<FAVE_API_ENDPOINT> -collection=<COLLECTION_NAME>
```

Docker:
```
docker run \
    -p 8080:8080 \
    w3ki-server -fave=<FAVE_API_ENDPOINT> -collection=<COLLECTION_NAME>
```

This will start a local http-serve which will serve wikipedia content on port `:8080`. 

#### How Indexer works:

Indexer uses [FaVe](https://github.com/fairDataSociety/FaVe) to store content and index them for doing semantic search. It is sanitizing the Article content and vectorizing it to do FaVe Nearest Neighbour Search. 

#### How Server works:

Server lists all the items in the server with "text/html" mimetype. 

For showing the content it checks "GET" request on "/wiki/XXXX". it reads the "XXXX" relative url, then finds the entry for that item in the FaVe/FairOS-dfs document store,
downloads the content and sends it back as response.

this project [uses code](https://github.com/akhenakh/gozim/blob) that is [MIT licensed](https://github.com/akhenakh/gozim/blob/master/LICENSE)