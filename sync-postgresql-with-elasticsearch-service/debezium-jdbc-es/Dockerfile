# Resolve the Elasticsearch sink connector (11.1.3, targets ES 7.x) and its full
# runtime dependency tree via Maven instead of hand-picking jars.
FROM maven:3.8-eclipse-temurin-11 AS es-connector-deps
COPY es-connector-pom.xml /build/pom.xml
RUN mvn -f /build/pom.xml dependency:copy-dependencies \
        -DoutputDirectory=/out -DincludeScope=runtime

FROM debezium/connect:2.6
ENV KAFKA_CONNECT_JDBC_DIR=$KAFKA_CONNECT_PLUGINS_DIR/kafka-connect-jdbc \
    KAFKA_CONNECT_ES_DIR=$KAFKA_CONNECT_PLUGINS_DIR/kafka-connect-elasticsearch

ARG POSTGRES_VERSION=42.2.8
ARG KAFKA_JDBC_VERSION=5.3.1

# Deploy PostgreSQL JDBC Driver
RUN cd /kafka/libs && curl -sO https://jdbc.postgresql.org/download/postgresql-$POSTGRES_VERSION.jar

# Deploy Kafka Connect JDBC
RUN mkdir $KAFKA_CONNECT_JDBC_DIR && cd $KAFKA_CONNECT_JDBC_DIR &&\
	curl -sO http://packages.confluent.io/maven/io/confluent/kafka-connect-jdbc/$KAFKA_JDBC_VERSION/kafka-connect-jdbc-$KAFKA_JDBC_VERSION.jar

# Deploy Confluent Elasticsearch sink connector
RUN mkdir $KAFKA_CONNECT_ES_DIR
COPY --from=es-connector-deps /out/ $KAFKA_CONNECT_ES_DIR/