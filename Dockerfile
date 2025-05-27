ARG GO_VERSION=1.23.2
FROM golang:${GO_VERSION} AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o main /app/main.go

# ========================================================================================

FROM scratch AS final

COPY --from=build /app/main /

CMD [ "/main" ]
