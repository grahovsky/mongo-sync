ARG IMAGE=golang:1.21
FROM ${IMAGE} AS build
LABEL stage=build

ENV BIN_FILE=mongo-sync

WORKDIR /workspace

COPY go.mod go.sum ./

ARG GOPROXY=https://proxy.golang.org,direct
RUN export GOPROXY=$GOPROXY

RUN go mod download -x

COPY . .

ARG LDFLAGS='-X main.release="develop"'
RUN CGO_ENABLED=0 GOOS=linux go build -v \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} cmd/*.go


FROM scratch AS final
LABEL stage=final

ENV BIN_FILE=mongo-sync
ENV CONFIG_FILE=/configs/config.yaml

USER 1000

WORKDIR /

COPY --from=build /workspace/${BIN_FILE} . 
COPY configs/config_example.yaml ${CONFIG_FILE}

ENTRYPOINT [ "/mongo-sync" ]
