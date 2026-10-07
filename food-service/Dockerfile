FROM eclipse-temurin:17-jdk AS builder

WORKDIR /app

COPY mvnw pom.xml ./
COPY .mvn ./.mvn
RUN chmod +x mvnw && ./mvnw -B -q dependency:go-offline

COPY src ./src
RUN ./mvnw -B -q package -DskipTests && cp target/food-service-*.jar app.jar

FROM eclipse-temurin:17-jre-alpine

WORKDIR /app

RUN adduser -D -u 10001 app

COPY --from=builder /app/app.jar app.jar

USER app

EXPOSE 8080

ENTRYPOINT ["java", "-jar", "app.jar"]
