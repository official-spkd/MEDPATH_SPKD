# Image produksi backend MedPath (dipakai Railway). Build statis, jalan di Alpine kecil.
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /medpath-api .

FROM alpine:3.23
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Jakarta
COPY --from=build /medpath-api /usr/local/bin/medpath-api
EXPOSE 8080
CMD ["medpath-api"]
