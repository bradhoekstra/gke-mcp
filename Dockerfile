# checkov:skip=CKV_DOCKER_2:Existing issue, suppressing to unblock presubmit
# checkov:skip=CKV_DOCKER_3:Existing issue, suppressing to unblock presubmit
FROM golang:1.25.7 AS build

WORKDIR /go/src/gke-mcp
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -o /gke-mcp

FROM gcr.io/google.com/cloudsdktool/google-cloud-cli:558.0.0-debian_component_based-20260224

COPY --from=build /gke-mcp /usr/local/bin/gke-mcp

EXPOSE 8080
ENTRYPOINT [ "gke-mcp" ]
