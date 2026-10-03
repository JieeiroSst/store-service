# Build: `sbt stage` produces a self-contained Play distribution.
FROM sbtscala/scala-sbt:eclipse-temurin-jammy-11.0.22_7_1.9.9_2.13.13 AS build
WORKDIR /src
# Dependencies first so they stay cached until the build definition changes.
COPY build.sbt ./
COPY project/build.properties project/plugins.sbt project/
RUN sbt -batch update
COPY app app
COPY conf conf
COPY public public
RUN sbt -batch stage

FROM eclipse-temurin:11-jre-jammy
RUN groupadd -g 10001 app && useradd -u 10001 -g app -M -d /app -s /usr/sbin/nologin app
COPY --from=build --chown=10001:10001 /src/target/universal/stage /app
USER 10001
WORKDIR /app
EXPOSE 9000
ENV JAVA_OPTS="-XX:MaxRAMPercentage=75 -Djava.io.tmpdir=/tmp"
# No RUNNING_PID file: the root filesystem is read-only in the cluster.
ENTRYPOINT ["/app/bin/book-service", "-Dpidfile.path=/dev/null", "-Dhttp.port=9000", "-Dlogger.resource=logback-container.xml"]
