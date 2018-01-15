FROM java:8-jre
MAINTAINER dingding <17212010004@fudan.edu.cn>

ADD ./target/test-backend-1.0.jar /app/

CMD ["java", "-Xmx200m", "-jar", "/app/test-backend-1.0.jar"]

EXPOSE 5001