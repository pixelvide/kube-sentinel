
# Build glab in a separate stage for caching
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS glab-builder
ARG TARGETOS TARGETARCH
RUN apk add --no-cache git
RUN git clone --depth 1 --branch v1.80.4 https://gitlab.com/gitlab-org/cli.git /tmp/glab && \
    cd /tmp/glab && \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=v1.80.4 -X main.commit=$(git rev-parse --short HEAD)" -o /go/bin/glab ./cmd/glab

# Build aws-iam-authenticator from source for multi-arch support
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS aws-iam-authenticator-builder
ARG TARGETOS TARGETARCH
RUN apk add --no-cache git
RUN git clone --depth 1 --branch v0.7.10 https://github.com/kubernetes-sigs/aws-iam-authenticator.git /tmp/aws-iam-authenticator && \
    cd /tmp/aws-iam-authenticator && \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /go/bin/aws-iam-authenticator ./cmd/aws-iam-authenticator

# Download helm client
FROM --platform=$BUILDPLATFORM alpine:3.19 AS helm-builder
ARG TARGETOS TARGETARCH
ENV HELM_VERSION=v3.17.0
RUN apk add --no-cache curl tar && \
    curl -fsSL https://get.helm.sh/helm-${HELM_VERSION}-${TARGETOS}-${TARGETARCH}.tar.gz -o helm.tar.gz && \
    tar -zxvf helm.tar.gz && \
    mv ${TARGETOS}-${TARGETARCH}/helm /usr/local/bin/helm && \
    chmod +x /usr/local/bin/helm

FROM node:20-alpine AS frontend-builder

WORKDIR /app/ui

COPY ui/package.json ui/pnpm-lock.yaml ./

RUN npm install -g pnpm && \
    pnpm install --frozen-lockfile

COPY ui/ ./
RUN pnpm run build

FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS backend-builder
ARG TARGETOS TARGETARCH

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

COPY --from=frontend-builder /app/static ./static
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o cloud-sentinel-k8s .

FROM gcr.io/distroless/static

WORKDIR /app

COPY --from=backend-builder /app/cloud-sentinel-k8s .
COPY --from=glab-builder /go/bin/glab /usr/local/bin/glab
COPY --from=aws-iam-authenticator-builder /go/bin/aws-iam-authenticator /usr/local/bin/aws-iam-authenticator
COPY --from=helm-builder /usr/local/bin/helm /usr/local/bin/helm

EXPOSE 8080

CMD ["./cloud-sentinel-k8s"]
