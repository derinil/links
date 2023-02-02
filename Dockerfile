FROM gcr.io/distroless/static:nonroot

COPY /bin/links /

EXPOSE 8080

CMD ["/links"]
