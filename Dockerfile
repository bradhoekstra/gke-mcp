FROM golang:1.24 AS build

WORKDIR /go/src/gke-mcp
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -o /gke-mcp
RUN tar czf /gke-mcp.tar.gz go.* cmd/ pkg/

FROM gcr.io/distroless/static-debian12
COPY --from=build /gke-mcp /
COPY --from=build /gke-mcp.tar.gz /go/src/
COPY . /go/src/gke-mcp

EXPOSE 3000
CMD [ "/gke-mcp" ]
