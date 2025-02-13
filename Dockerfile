# Build the urlshortener binary
FROM golang:1.24

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY pkg/ pkg/
COPY cmd/ cmd/
COPY main.go main.go

# Build
RUN go build -o /bin/meetingepd ./main.go

ENTRYPOINT ["/bin/meetingepd"]
CMD [ "serve" ]

EXPOSE     8099
EXPOSE     50051
