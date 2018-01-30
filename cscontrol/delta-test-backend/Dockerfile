FROM java:8-jre

ADD ./target/delta-test-backend-1.0.jar /app/

CMD ["java", "-Xmx200m", "-jar", "/app/delta-test-backend-1.0.jar"]

EXPOSE 5002