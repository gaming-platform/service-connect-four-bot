FROM golang:1.25-alpine AS development

WORKDIR /project

RUN go install github.com/air-verse/air@v1.63.4

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /usr/local/bin/connect-four-bot .

CMD ["air", \
    "--build.cmd", \
    "go build -o /usr/local/bin/connect-four-bot .", \
    "--build.entrypoint", \
    "/usr/local/bin/connect-four-bot"]

FROM scratch

COPY --from=development /usr/local/bin/connect-four-bot /usr/local/bin/connect-four-bot

CMD ["connect-four-bot"]
