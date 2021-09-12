FROM golang
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN apt update -y && apt install make
RUN mkdir /exporter
ADD . /exporter/
WORKDIR /exporter
RUN make build
EXPOSE 9001
CMD ./miwifi_exporter