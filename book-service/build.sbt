lazy val root = (project in file("."))
  .enablePlugins(PlayScala)
  .settings(
    name := """book-service""",
    version := "1.0-SNAPSHOT",
    scalaVersion := "2.13.10",
    libraryDependencies ++= Seq(
      guice,
      jdbc,
      evolutions,
      "org.playframework.anorm" %% "anorm" % "2.6.10",
      "org.postgresql" % "postgresql" % "42.7.4",
      // In-memory database for tests and `sbt run` (PostgreSQL mode).
      "com.h2database" % "h2" % "1.4.199",
      // Book content lives in MinIO; the AWS SDK speaks S3 to it. The JDK
      // HTTP client replaces the default Netty/Apache ones.
      ("software.amazon.awssdk" % "s3" % "2.25.70")
        .exclude("software.amazon.awssdk", "netty-nio-client")
        .exclude("software.amazon.awssdk", "apache-client"),
      "software.amazon.awssdk" % "url-connection-client" % "2.25.70",
      "org.scalatestplus.play" %% "scalatestplus-play" % "5.0.0" % Test
    ),
    // The deployable is the `stage` distribution; skip Scaladoc (it would also
    // fail on -Xfatal-warnings).
    Compile / doc / sources := Seq.empty,
    Compile / packageDoc / publishArtifact := false,
    scalacOptions ++= Seq(
      "-feature",
      "-deprecation",
      "-Xfatal-warnings"
    ),
    dependencyOverrides ++= Seq(
      "com.google.inject" % "guice" % "5.1.0",
      "com.google.inject.extensions" % "guice-assistedinject" % "5.1.0"
    )
  )
