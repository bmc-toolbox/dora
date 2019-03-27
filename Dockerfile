# Build
FROM golang:1 as build

WORKDIR /build
ADD . .

# Tests
RUN go test -tags="gingonic" -mod=vendor ./...

RUN GOOS=linux go build -mod=vendor -tags="gingonic" -a -installsuffix cgo -o dora .

# Run
FROM centos

RUN adduser -s /bin/false dora

RUN mkdir -p /etc/dora
COPY dora-simple.yaml /etc/bmc-toolbox/dora.yaml
COPY kea-simple.conf /etc/kea/kea.conf

COPY --from=build /build/dora /usr/bin

EXPOSE 8000
USER dora

ENTRYPOINT ["/usr/bin/dora"]
