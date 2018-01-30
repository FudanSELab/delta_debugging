FROM java:8-jre

ADD ./target/api-server-1.0.jar /app/
CMD ["java", "-Xmx200m", "-jar", "/app/api-server-1.0.jar"]

EXPOSE 18898
