FROM java:8-jre
MAINTAINER Li Wenhai <16212010016@fudan.edu.cn>

ADD ./target/api-server-1.0.jar /app/
CMD ["java", "-Xmx200m", "-jar", "/app/api-server-1.0.jar"]

EXPOSE 18898
