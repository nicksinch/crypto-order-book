FROM golang:1.22-alpine

WORKDIR /go/src/app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o "/ob" "./go/cmd"

EXPOSE 8080

CMD [ "/ob" ]