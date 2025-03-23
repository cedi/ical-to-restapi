FROM alpine:3.7

LABEL org.opencontainers.image.source="https://github.com/SpechtLabs/CalendarAPI"
LABEL org.opencontainers.image.description="CalendarAPI is a service that parses iCal files and exposes their content via gRPC or a REST API."
LABEL org.opencontainers.image.licenses="MIT"

COPY ./calenderapi /bin/calenderapi

ENTRYPOINT ["/bin/meetingepd"]
CMD [ "serve" ]

EXPOSE     8099
EXPOSE     50051
