# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.21-alpine3.18 AS build
ARG TARGETOS TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    GOOS=$TARGETOS GOARCH=$TARGETARCH go build -v -o "/usr/local/bin/terraform-backend-github" "github.com/andreygubarev/terraform-backend-github"

FROM alpine:3.18 AS terraform-backend-github
COPY --from=build /usr/local/bin/terraform-backend-github /usr/local/bin/terraform-backend-github

ENV GIN_MODE=release
ENV GITHUB_TOKEN=

EXPOSE 8080/tcp
CMD ["terraform-backend-github"]
