ARG IMAGE=golang:1.20
FROM ${IMAGE} AS build
LABEL stage=build

WORKDIR /workspace

COPY go.mod go.sum ./

ARG GOPROXY=https://proxy.golang.org,direct
RUN export GOPROXY=$GOPROXY

RUN go mod download

COPY main.go .

RUN CGO_ENABLED=0 GOOS=linux go build -a -o mongo-sync

FROM scratch AS final
LABEL stage=final

USER 1000

WORKDIR /

COPY --from=build /workspace/mongo-sync . 
COPY config_example ./config

ENV CONFIG_PATH=

CMD [ "./mongo-sync" ]