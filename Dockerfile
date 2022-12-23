FROM public.ecr.aws/docker/library/golang:1.18

RUN apt update && apt install -y --no-install-recommends \
  git=1:2.30.2-1 \
  make=4.3-4.1 \
  zip=3.0-12

WORKDIR /builddir

COPY go.mod go.sum ./
COPY *.go ./

RUN go mod tidy
