## zwi-fave

This project uses FaVe and fairdatasociety/huggingface-vectorizer to index, store and search zwi files content.

It has two components

#### Uploader: 

[Uploader](./cmd/uploader/README.md) can read [ZWI](https://docs.encyclosphere.org/#/zwi-format) format and upload data to FaVe, ultimately storing content on [swarm](https://www.ethswarm.org/).

#### Server:

Server will start a http-server that can fetch content from FaVe and display them in the web browser.  

### How to serve?

```
cd cmd/server
go build
```

Docker:
```
docker build -f  Dockerfile.server --tag zwi-fave-server .
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
docker run zwi-fave-server -h
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
    zwi-fave-server -fave=<FAVE_API_ENDPOINT> -collection=<COLLECTION_NAME>
```

This will start a local http-serve which will serve wikipedia content on port `:8080`.

this project [uses code](https://github.com/akhenakh/gozim/blob) that is [MIT licensed](https://github.com/akhenakh/gozim/blob/master/LICENSE)

### Docker Compose

Best way to run the server is to use a single docker-compose file that will start vectorizer, FaVe and the server.

Copy the following docker-compose file and save it as `docker-compose.yml` in a directory.
```
version: '3'
services:
  vectorizer:
    command:
      - --model-name
      - sentence-transformers/all-mpnet-base-v2
    image: fairdatasociety/huggingface-vectorizer:latest
    ports:
      - 9876:9876
    restart: on-failure:0

  fave:
    command:
      - --host
      - 0.0.0.0
      - --port
      - '1234'
      - --write-timeout
      - 1500m
      - --read-timeout
      - 1500m
    image: fairdatasociety/fave:latest
    ports:
      - 1234:1234
    restart: on-failure:0
    environment:
      BEE_API: <BEE_URL>
      RPC_API: <RPC_ENDPOINT_FOR_ENS>
      STAMP_ID: 0
      USER: <USER_NAME>
      PASSWORD: <PASSWORD>
      POD: <POD_NAME>
      VECTORIZER_URL: http://vectorizer:9876
      VERBOSE: true
  zwi-fave:
    command:
      - -collection
      - <COLLECTION_NAME>
      - -port
      - '8526'
      - -fave
      - http://fave:1234/v1
    image: fairdatasociety/zwi-fave:latest
    ports:
      - 8526:8526
    restart: on-failure:0
```

Change the environment variables to match your setup.
```
BEE_API: <BEE_URL> # Bee api endpoint (Make sure to use local ip (e.g. 192.168.x.x), not "localhost" or "127.0.0.1")
RPC_API: <RPC_ENDPOINT_FOR_ENS> # This is for fairOS-dfs to authenticate with ENS
USER: <USER_NAME> # FairOS-dfs user name
PASSWORD: <PASSWORD> # FairOS-dfs password
POD: <POD_NAME> # FairOS-dfs pod name to store content
COLLECTION_NAME: <COLLECTION_NAME> # Collection name to store content on FaVe
```

Then run `docker-compose up` to start the server. The server should be available on port `:8526`.
