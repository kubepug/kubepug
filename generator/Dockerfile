FROM cgr.dev/chainguard/go:latest as build

WORKDIR /work
COPY main.go /work
RUN go mod init generator; go mod tidy; go build -o generator .

FROM cgr.dev/chainguard/go:latest

COPY --from=build /work/generator /generator
COPY <<EOF build.sh
#!/bin/sh
mkdir -p /data/apis 
cd /data/
go mod init generator
cd /data/apis
GOPATH=\$(pwd) go get k8s.io/api@\${VERSION}
GOPATH=\$(pwd) /generator -i k8s.io/api/./...
EOF

ENV VERSION=latest
ENV GOPATH=/work
ENTRYPOINT ["/bin/sh", "/build.sh"]