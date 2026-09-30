FROM golang:1.18
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /usr/local/bin/drive-tree

# Listen on all interfaces so the port can be published to the host
ENTRYPOINT ["drive-tree", "-addr", ":8080", "-no-browser"]

# docker build -t my-drive-tree-app .
# docker run -it --rm -p 8080:8080 --name my-running-app my-drive-tree-app
