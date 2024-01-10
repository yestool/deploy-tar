Use this tool to deploy any static website (like Hugo, Hexo, Astro, Jekyll, VuePress.... ) with a VPS. 

## client

```
# build client
go build -o ./client/bin/deploy-tar ./client/main.go
```


### client config

deploy-tar config is a yaml file

```
apiKey: 1111111111
server: http://127.0.0.1:8081/upload
webPath: /www/wwwroot/test123.com/
tarPath: /root/go-code/deploy-tar/test/
webSite: test123.com
```


```
> deplay-tar --config .deploy-tar.yaml 

apiKey: use apikey to secure your system.

server: the server upload handler location.

webPath: Untargz to this directory, if you use docker, just volumes your real websit directory to this directory.

tarPath: can be **xxx.tar.gz** or a **directory**. if it's a directory, Package as tar.gz file first,then upload.

webSite: just a websit id tag.
```




## server


```
# build server

go build -o ./server/bin/deploy-tar-server ./server/main.go

# build dockerfile
docker build --no-cache -t yestool/deploy-tar:v1 .
```


### use docker-compose 


```
version: '3'
services:
  deploy-tar:
    image: yestool/deploy-tar:v1.0
    ports:
      - "8081:8080"
    environment:
      APP_APIKEY: 11111111
    volumes:
      - ./tars:/uploadfiles
      - ./webs:/www/wwwroot
```

### server config

default config:

```
apiKey: 123456
serverPort: 8080
keepFiles: 3
maxUploadSize: 104857600
funcHandle: /upload
```

you can use ENV cover , ENV prefix is **APP**.

### Nginx config

Hide the server behind websit location:


```
location /upxyz123/ {
  client_max_body_size  1024m;
  proxy_set_header Host $host:$server_port;
  proxy_set_header X-Real-Ip $remote_addr;
  proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
  proxy_http_version 1.1;
  proxy_set_header Upgrade $http_upgrade;
  proxy_set_header Connection "upgrade";
  proxy_connect_timeout 99999;
  proxy_pass http://127.0.0.1:8081/;
}
```

in this config, the client config **server** is : **https://xxxx.com/upxyz123/upload**