FROM java:8-jre

ADD ./target/delta-backend-1.0.jar /app/

CMD ["java", "-Xmx200m", "-jar", "/app/delta-backend-1.0.jar"]

EXPOSE 5002