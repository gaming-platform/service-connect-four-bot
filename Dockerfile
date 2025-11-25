FROM golang:1.25-alpine AS development

WORKDIR /project

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /usr/local/bin/connect-four-bot .

CMD ["connect-four-bot"]

FROM scratch

COPY --from=development /usr/local/bin/connect-four-bot /usr/local/bin/connect-four-bot

CMD ["connect-four-bot"]
