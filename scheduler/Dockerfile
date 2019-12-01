FROM golang:latest
RUN mkdir /app
ADD . /app/
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scheduler .

RUN ["chmod", "+x", "/app/scheduler"]
CMD ["/app/scheduler"]