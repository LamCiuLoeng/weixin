FROM golang:onbuild
RUN mkdir /code
ADD . /code
WORKDIR /code
RUN go build  -o main  app.go
CMD ["./main"]
EXPOSE 5000
