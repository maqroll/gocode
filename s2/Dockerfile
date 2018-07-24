FROM alpine
WORKDIR /bin
# compile with CGO_ENABLED=0
ADD main /bin
EXPOSE 8081
RUN chmod +x /bin/main
CMD /bin/main
