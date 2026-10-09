# Resolve the Elasticsearch sink connector (11.1.3, targets ES 7.x) and its full
# runtime dependency tree via Maven instead of hand-picking jars.
FROM maven:3.8-eclipse-temurin-11 AS es-connector-deps
COPY es-connector-pom.xml /build/pom.xml
RUN mvn -f /build/pom.xml dependency:copy-dependencies \
        -DoutputDirectory=/out -DincludeScope=runtime

# debezium/connect already ships the Debezium MongoDB source connector.
FROM debezium/connect:2.6
ENV KAFKA_CONNECT_ES_DIR=$KAFKA_CONNECT_PLUGINS_DIR/kafka-connect-elasticsearch

# Deploy Confluent Elasticsearch sink connector
RUN mkdir $KAFKA_CONNECT_ES_DIR
COPY --from=es-connector-deps /out/ $KAFKA_CONNECT_ES_DIR/
