# w3kipedia uploader

This is a demo for Uploading wikipedia zim files on FaVeDB. FaVeDB internally uses fariOS-dfs => swarm.

## How does it work?

### What do we need?

- First of all we need [zwi files](https://docs.encyclosphere.org/#/zwi-format) to upload. FaVe only supports English language for now.
- We need a running [FaVeDB](https://github.com/fairDataSociety/FaVe) server.

### Steps

- Uploader create the collection on FaVe
- It reads the zwi files
- Tokenizes the article content
- Processes the documents to be stored on FaVe
- We also process other files (js, css, images, etc) on FaVe without indexing or vectorization
- uploads the documents on FaVe

## How to run?

### Build
On Ubuntu/Debian :
```
cd cmd/uploader
go build
```

On OSX:
```
cd cmd/uploader
CGO_CFLAGS=`pkg-config --cflags liblzma` go build 
```

Docker:
```
docker build -f Dockerfile.uploader --tag w3ki-uploader .
```
#### Help :

```
./uploader --help
Usage of ./uploader:
  -fave string
        FaVe API endpoint ("http://localhost:1234/v1")
  -collection string
        Collection name to store on FaVe
  -zwi string
        path of the folder that contains the zwi files
```

Docker:
```
docker run w3ki-uploader -h
```


#### Running :

Binary:
```
./uploader -fave=<FAVE_API_ENDPOINT> -collection=<COLLECTION_NAME> -zwi=zwiPath
```

Docker:
```
docker run \
    -v <PATH_TO_ZWI>:/go/src/github.com/onepeerlabs/w3kipedia/<ZWI_FILE_PATH> \
    w3ki-uploader -zwi=<ZWI_FILE_PATH> -fave=<FAVE_API_ENDPOINT> -collection=<COLLECTION_NAME>
```