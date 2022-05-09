FROM golang:1.18
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /usr/local/bin/drive-tree

RUN mkdir /drive-tree
WORKDIR /drive-tree
ENTRYPOINT ["/bin/bash"]

# docker build -t my-drive-tree-app .
# docker run -it --rm -v $(pwd):/drive-tree -p 8080:8080 --name my-running-app my-drive-tree-app